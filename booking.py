from datetime import datetime, timedelta, timezone
from urllib.parse import urlencode
from mailer import send_booking_invite

from flask import Blueprint, jsonify, render_template, request, url_for

from database import Lead, db, get_booking_settings
from google_calendar import create_calendar_event, get_busy_events, google_calendar_available
from lead_score import calculate_lead_score

booking_bp = Blueprint('booking', __name__)
SAST = timezone(timedelta(hours=2), 'SAST')

@booking_bp.route('/book')
def book():
    return render_template('book.html')


def parse_booking_time(value):
    if not value:
        return None

    parsed = datetime.fromisoformat(value)
    if parsed.tzinfo is None:
        parsed = parsed.replace(tzinfo=SAST)
    return parsed


def google_calendar_url(start_time, name, company, email, phone, project_type, notes):
    end_time = start_time + timedelta(minutes=30)
    start_utc = start_time.astimezone(timezone.utc).strftime('%Y%m%dT%H%M%SZ')
    end_utc = end_time.astimezone(timezone.utc).strftime('%Y%m%dT%H%M%SZ')

    details = "\n".join([
        "Discovery call booked through JO4 Dev.",
        f"Name: {name}",
        f"Company: {company or 'Not provided'}",
        f"Email: {email}",
        f"Phone: {phone or 'Not provided'}",
        f"Project type: {project_type}",
        "",
        notes or "No extra notes provided."
    ])

    params = {
        "action": "TEMPLATE",
        "text": f"JO4 Dev Discovery Call - {company or name}",
        "dates": f"{start_utc}/{end_utc}",
        "details": details,
        "location": "Google Meet / Phone call",
        "trp": "false",
    }

    return f"https://calendar.google.com/calendar/render?{urlencode(params)}"


def parse_time(value, fallback):
    try:
        hour, minute = [int(part) for part in value.split(':', 1)]
        return hour, minute
    except (AttributeError, TypeError, ValueError):
        return fallback


def local_day_bounds(date_value):
    day = datetime.fromisoformat(date_value).date()
    return (
        datetime(day.year, day.month, day.day, 0, 0, tzinfo=SAST),
        datetime(day.year, day.month, day.day, 23, 59, 59, tzinfo=SAST)
    )


def overlaps(start_time, end_time, busy_start, busy_end):
    return start_time < busy_end and end_time > busy_start


def build_available_slots(settings, date_value):
    day_start, day_end = local_day_bounds(date_value)
    start_hour, start_minute = parse_time(settings.workday_start, (9, 0))
    end_hour, end_minute = parse_time(settings.workday_end, (17, 0))
    work_start = day_start.replace(hour=start_hour, minute=start_minute, second=0)
    work_end = day_start.replace(hour=end_hour, minute=end_minute, second=0)

    duration = timedelta(minutes=settings.meeting_duration_minutes or 30)
    step = timedelta(minutes=settings.slot_step_minutes or 30)
    buffer_time = timedelta(minutes=settings.buffer_minutes or 30)
    min_start = datetime.now(SAST) + timedelta(hours=settings.min_notice_hours or 0)

    events, calendar_error = get_busy_events(settings, day_start, day_end)
    blocked = [
        (
            event["start"].astimezone(SAST) - buffer_time,
            event["end"].astimezone(SAST) + buffer_time
        )
        for event in events
    ]

    slots = []
    cursor = work_start
    while cursor + duration <= work_end:
        slot_end = cursor + duration
        blocked_by_event = any(overlaps(cursor, slot_end, busy_start, busy_end) for busy_start, busy_end in blocked)
        if cursor >= min_start and not blocked_by_event and cursor.weekday() < 5:
            slots.append({
                "value": cursor.isoformat(),
                "label": cursor.strftime('%H:%M'),
                "ends_at": slot_end.isoformat()
            })
        cursor += step

    return {
        "date": date_value,
        "slots": slots,
        "busy": [
            {
                "summary": event["summary"],
                "start": event["start"].astimezone(SAST).isoformat(),
                "end": event["end"].astimezone(SAST).isoformat()
            }
            for event in events
        ],
        "calendar_connected": calendar_error is None,
        "calendar_error": calendar_error,
        "settings": {
            "duration": settings.meeting_duration_minutes or 30,
            "buffer": settings.buffer_minutes or 30,
            "horizon_days": settings.booking_horizon_days or 21,
            "workday_start": settings.workday_start,
            "workday_end": settings.workday_end
        }
    }


@booking_bp.route('/api/booking-availability')
def booking_availability():
    date_value = request.args.get('date') or datetime.now(SAST).date().isoformat()
    settings = get_booking_settings()

    try:
        availability = build_available_slots(settings, date_value)
    except ValueError:
        return jsonify({"status": "error", "message": "Invalid date."}), 400

    availability["status"] = "success"
    availability["google_packages_installed"] = google_calendar_available()
    return jsonify(availability)


@booking_bp.route('/api/book-call', methods=['POST'])
def book_call():
    data = request.get_json(silent=True) or {}

    name = (data.get('name') or '').strip()
    email = (data.get('email') or '').strip()
    company = (data.get('company') or '').strip()
    phone = (data.get('phone') or '').strip()
    project_type = (data.get('project_type') or 'static').strip()
    notes = (data.get('notes') or '').strip()
    scheduled_at_raw = (data.get('scheduled_at') or '').strip()

    if not name or not email or not scheduled_at_raw:
        return jsonify({
            "status": "error",
            "message": "Name, email, and a meeting time are required."
        }), 400

    try:
        scheduled_at = parse_booking_time(scheduled_at_raw)
    except ValueError:
        return jsonify({
            "status": "error",
            "message": "Please choose a valid meeting time."
        }), 400

    if scheduled_at <= datetime.now(SAST):
        return jsonify({
            "status": "error",
            "message": "Please choose a future meeting time."
        }), 400

    settings = get_booking_settings()
    availability = build_available_slots(settings, scheduled_at.date().isoformat())
    valid_slot_values = {slot["value"] for slot in availability["slots"]}
    
    if scheduled_at.isoformat() not in valid_slot_values:
        return jsonify({
            "status": "error",
            "message": "That time is no longer available. Please choose another slot."
        }), 409

    scoring_input = {
        "client_name": name,
        "client_company": company,
        "contact_role": data.get('contact_role') or 'Owner',
        "budget": data.get('budget') or 0,
        "project_type": project_type,
        "visited_pages": ['/book'],
        "whatsapp_engagement": "replied" if phone else "read",
        "phone_number": phone,
        "last_activity_date": datetime.utcnow().strftime('%Y-%m-%d')
    }
    score_result = calculate_lead_score(scoring_input)

    lead = Lead(
        client_name=name,
        client_company=company,
        project_type=project_type,
        budget=float(data.get('budget') or 0),
        contact_role=data.get('contact_role') or 'Owner',
        phone_number=phone,
        whatsapp_engagement='replied' if phone else 'read',
        target_project=project_type,
        score=int(round(score_result['score'])),
        explicit_score=score_result['breakdown'].get('explicit'),
        implicit_score=score_result['breakdown'].get('implicit'),
        urgency_score=score_result['breakdown'].get('urgency'),
        status='New',
        last_activity_date=datetime.utcnow()
    )

    try:
        db.session.add(lead)
        db.session.commit()
    except Exception as e:
        db.session.rollback()
        return jsonify({
            "status": "error",
            "message": f"Booking could not be saved: {e}"
        }), 500

    try:
        calendar_event, calendar_error = create_calendar_event(settings, scheduled_at, {
            "name": name,
            "email": email,
            "company": company,
            "phone": phone,
            "project_type": project_type,
            "notes": notes
        })
        
        # ---------------------------------------------------------
        # NEW: Send the automated Brevo email with the .ics invite
        # ---------------------------------------------------------
        if calendar_event:
            # Note: We are passing 'scheduled_at' directly as a datetime object!
            send_booking_invite(email, name, scheduled_at)
            
    except Exception as e:
        calendar_event = None
        calendar_error = str(e)
        
    fallback_calendar_url = google_calendar_url(
        scheduled_at,
        name,
        company,
        email,
        phone,
        project_type,
        notes
    )

    return jsonify({
        "status": "success",
        "message": "Your call is booked." if calendar_event else "Your request is saved. Calendar automation still needs Google credentials.",
        "calendar_url": calendar_event.get('htmlLink') if calendar_event else fallback_calendar_url,
        "calendar_event_created": calendar_event is not None,
        "calendar_error": calendar_error,
        "lead_id": lead.id,
        "score": score_result['score'],
        "classification": score_result['classification'],
        "admin_url": url_for('admin.leads_page')
    })