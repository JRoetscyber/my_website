from datetime import datetime, timedelta, timezone
import os
import uuid


SCOPES = ['https://www.googleapis.com/auth/calendar']


def google_calendar_available():
    try:
        import google.oauth2.service_account  # noqa: F401
        import googleapiclient.discovery  # noqa: F401
        return True
    except ImportError:
        return False


def get_calendar_service(settings):
    if not google_calendar_available():
        return None, "Google Calendar packages are not installed."

    credential_path = settings.service_account_file or os.getenv('GOOGLE_SERVICE_ACCOUNT_FILE')
    calendar_id = settings.calendar_id or os.getenv('GOOGLE_CALENDAR_ID') or 'primary'

    if not credential_path:
        return None, "Google service account file is not configured."

    if not os.path.isabs(credential_path):
        credential_path = os.path.join(os.getcwd(), credential_path)

    if not os.path.exists(credential_path):
        return None, f"Google service account file not found: {credential_path}"

    from google.oauth2 import service_account
    from googleapiclient.discovery import build

    credentials = service_account.Credentials.from_service_account_file(
        credential_path,
        scopes=SCOPES
    )
    service = build('calendar', 'v3', credentials=credentials, cache_discovery=False)
    return service, calendar_id


def get_busy_events(settings, start_time, end_time):
    service, calendar_or_error = get_calendar_service(settings)
    if not service:
        return [], calendar_or_error

    response = service.events().list(
        calendarId=calendar_or_error,
        timeMin=start_time.astimezone(timezone.utc).isoformat(),
        timeMax=end_time.astimezone(timezone.utc).isoformat(),
        singleEvents=True,
        orderBy='startTime'
    ).execute()

    events = []
    for item in response.get('items', []):
        start = item.get('start', {})
        end = item.get('end', {})
        start_value = start.get('dateTime') or start.get('date')
        end_value = end.get('dateTime') or end.get('date')
        if not start_value or not end_value:
            continue

        events.append({
            "summary": item.get('summary', 'Busy'),
            "start": datetime.fromisoformat(start_value.replace('Z', '+00:00')),
            "end": datetime.fromisoformat(end_value.replace('Z', '+00:00')),
        })

    return events, None


def create_calendar_event(settings, start_time, client):
    service, calendar_or_error = get_calendar_service(settings)
    if not service:
        return None, calendar_or_error

    end_time = start_time + timedelta(minutes=settings.meeting_duration_minutes or 30)
    details = "\n".join([
        "Discovery call booked through JO4 Dev.",
        f"Name: {client['name']}",
        f"Company: {client.get('company') or 'Not provided'}",
        f"Email: {client['email']}",
        f"Phone: {client.get('phone') or 'Not provided'}",
        f"Project type: {client.get('project_type') or 'Not provided'}",
        "",
        client.get('notes') or "No extra notes provided."
    ])

    body = {
        "summary": f"JO4 Dev Discovery Call - {client.get('company') or client['name']}",
        "description": details,
        "start": {
            "dateTime": start_time.isoformat(),
            "timeZone": "Africa/Johannesburg"
        },
        "end": {
            "dateTime": end_time.isoformat(),
            "timeZone": "Africa/Johannesburg"
        },
        "reminders": {
            "useDefault": False,
            "overrides": [
                {"method": "email", "minutes": settings.reminder_minutes or 30},
                {"method": "popup", "minutes": settings.reminder_minutes or 30}
            ]
        }
    }

    body["location"] = "Google Meet / Phone call (Link will be sent separately)"

    event = service.events().insert(
        calendarId=calendar_or_error,
        body=body,
        sendUpdates='none',
        conferenceDataVersion=0
    ).execute()

    return event, None
