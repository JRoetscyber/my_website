from flask import Blueprint, render_template, request, redirect, url_for, flash, jsonify, make_response
from flask_login import login_required, logout_user
from database import db, Lead, Project, AutomationLog, Analytics, BlogPost, FAQ, BookingSettings, Transaction, Service, FAQSubmission, get_booking_settings, make_slug
from google_calendar import google_calendar_available
from datetime import datetime, timedelta, date
from sqlalchemy import func, text
from lead_score import calculate_lead_score
import os
import re
import unicodedata
import io, csv

def secure_filename(filename):
    filename = str(filename)
    filename = unicodedata.normalize('NFKD', filename).encode('ascii', 'ignore').decode('ascii')
    filename = filename.replace(os.path.sep, ' ')
    if os.path.altsep:
        filename = filename.replace(os.path.altsep, ' ')
    filename = re.sub(r'[^A-Za-z0-9_.-]+', '_', filename).strip('._')
    return filename or 'upload'


def save_blog_media(file):
    filename = secure_filename(file.filename)
    upload_dir = os.path.join('static', 'uploads', 'blog')
    os.makedirs(upload_dir, exist_ok=True)
    file_path = os.path.join(upload_dir, filename)
    file.save(file_path)
    return f"/{file_path.replace(os.sep, '/')}"


admin_bp = Blueprint('admin', __name__)


def normalize_lead_status(status):
    mapping = {
        'proposal sent': 'Negotiating',
        'proposal':      'Negotiating',
        # Lead classifications stored as status by the old /api/new-lead route
        'hot':           'New',
        'warm':          'New',
        'cold':          'New',
    }
    value = (status or 'New').strip()
    normalized = mapping.get(value.lower(), value)
    # Reject any value that isn't a valid pipeline stage
    valid = {'New', 'Contacted', 'Negotiating', 'Closed', 'Lost'}
    return normalized if normalized in valid else 'New'


def normalize_project_type(project_type):
    value = str(project_type or 'Static').strip()
    if value.lower() == 'website':
        return 'Static'
    return value


def serialize_lead(lead):
    return {
        "id": lead.id,
        "client_name": lead.client_name,
        "client_company": lead.client_company,
        "project_type": lead.project_type or lead.target_project,
        "budget": lead.budget,
        "contact_role": lead.contact_role,
        "phone_number": lead.phone_number,
        "whatsapp_engagement": lead.whatsapp_engagement,
        "target_project": lead.target_project,
        "score": lead.score,
        "status": normalize_lead_status(lead.status),
        "loss_reason": lead.loss_reason,
        "created_at": lead.created_at.isoformat() if lead.created_at else None,
        "last_activity_date": lead.last_activity_date.isoformat() if lead.last_activity_date else None,
        "breakdown": {
            "explicit": lead.explicit_score,
            "implicit": lead.implicit_score,
            "urgency": lead.urgency_score
        }
    }


def prepare_leads_for_display(leads):
    for lead in leads:
        lead.display_status = normalize_lead_status(lead.status or 'New')
        lead.display_project_type = lead.project_type or lead.target_project or 'Unspecified'
        lead.display_score = lead.score if lead.score is not None else None
    return leads


def safe_budget_value(value):
    try:
        return float(value or 0)
    except (TypeError, ValueError):
        return 0.0


def build_admin_context(active_page, selected_month=None, selected_year=None):
    data_error = None
    top_pages = []

    try:
        leads_list = Lead.query.order_by(Lead.created_at.desc(), Lead.id.desc()).all()
        leads_list = prepare_leads_for_display(leads_list)
        projects_list = Project.query.order_by(Project.deployed_at.desc()).all()
        blogs_list = BlogPost.query.order_by(BlogPost.created_at.desc()).all()
        faqs_list = FAQ.query.order_by(FAQ.display_order.asc(), FAQ.created_at.desc()).all()
        services_list = Service.query.order_by(Service.display_order.asc()).all()
        faq_submissions = FAQSubmission.query.order_by(FAQSubmission.created_at.desc()).all()
        booking_settings = get_booking_settings()
        
        transactions_query = Transaction.query
        if selected_month and selected_year:
            # Filter transactions for the selected month and year
            start_date = date(selected_year, selected_month, 1)
            # Calculate the last day of the month
            if selected_month == 12:
                end_date = date(selected_year + 1, 1, 1) - timedelta(days=1)
            else:
                end_date = date(selected_year, selected_month + 1, 1) - timedelta(days=1)
            transactions_query = transactions_query.filter(
                Transaction.date >= start_date,
                Transaction.date <= end_date
            )
        transactions_list = transactions_query.order_by(Transaction.date.desc(), Transaction.id.desc()).all()
        
        logs_list = AutomationLog.query.order_by(AutomationLog.timestamp.desc()).limit(10).all()

        total_income = sum(t.amount for t in transactions_list if t.type == 'Income')
        total_expenses = sum(t.amount for t in transactions_list if t.type == 'Expense')
        net_balance = total_income - total_expenses
        
        # Calculate Salary & Reinvestment (20/80 split of profit)
        salary_draw = max(0, net_balance * 0.20)
        reinvestment_fund = max(0, net_balance * 0.80)

        total_views = Analytics.query.count()
        one_week_ago = datetime.utcnow() - timedelta(days=7)
        scripts_this_week = AutomationLog.query.filter(AutomationLog.timestamp >= one_week_ago).count()

        total_leads = len(leads_list)
        closed_leads = sum(1 for lead in leads_list if lead.display_status == 'Closed')
        conversion_rate = (closed_leads / total_leads * 100) if total_leads > 0 else 0
        active_statuses = {'New', 'Contacted', 'Negotiating'}
        active_leads = [lead for lead in leads_list if lead.display_status in active_statuses]
        active_leads_count = len(active_leads)

        pipeline_forecast = sum(safe_budget_value(lead.budget) for lead in active_leads)
        scored_active_leads = [lead.display_score for lead in active_leads if lead.display_score is not None]
        average_lead_score = (sum(scored_active_leads) / len(scored_active_leads)) if scored_active_leads else 0

        funnel_counts = {
            'new': sum(1 for lead in leads_list if lead.display_status == 'New'),
            'contacted': sum(1 for lead in leads_list if lead.display_status == 'Contacted'),
            'negotiating': sum(1 for lead in leads_list if lead.display_status == 'Negotiating'),
            'closed': sum(1 for lead in leads_list if lead.display_status == 'Closed')
        }

        project_type_metrics = {}
        for lead in leads_list:
            key = lead.display_project_type
            bucket = project_type_metrics.setdefault(key, {
                'project_type': key,
                'total': 0,
                'closed': 0,
                'lost': 0,
                'open': 0,
                'conversion_rate': 0
            })
            bucket['total'] += 1
            if lead.display_status == 'Closed':
                bucket['closed'] += 1
            elif lead.display_status == 'Lost':
                bucket['lost'] += 1
            else:
                bucket['open'] += 1

        win_loss_by_project_type = sorted(
            [
                {
                    **bucket,
                    'conversion_rate': round((bucket['closed'] / bucket['total'] * 100), 1) if bucket['total'] else 0
                }
                for bucket in project_type_metrics.values()
            ],
            key=lambda item: (item['conversion_rate'], item['closed'], item['total']),
            reverse=True
        )

        top_pages_query = db.session.query(
            Analytics.page_path,
            func.count(Analytics.id).label('count')
        ).group_by(Analytics.page_path).order_by(func.count(Analytics.id).desc()).limit(5).all()

        max_count = top_pages_query[0].count if top_pages_query else 1
        for p in top_pages_query:
            top_pages.append({
                'path': p.page_path,
                'count': p.count,
                'percentage': (p.count / max_count * 100) if max_count > 0 else 0
            })
    except Exception as e:
        db.session.rollback()
        data_error = str(e)
        leads_list = []
        projects_list = []
        blogs_list = []
        faqs_list = []
        services_list = []
        faq_submissions = []
        booking_settings = BookingSettings()
        transactions_list = []
        logs_list = []
        total_views = 0
        scripts_this_week = 0
        conversion_rate = 0
        active_leads_count = 0
        pipeline_forecast = 0
        average_lead_score = 0
        total_income = 0
        total_expenses = 0
        net_balance = 0
        salary_draw = 0
        reinvestment_fund = 0
        funnel_counts = {'new': 0, 'contacted': 0, 'negotiating': 0, 'closed': 0}
        win_loss_by_project_type = []

    return {
        'leads': leads_list,
        'projects': projects_list,
        'blogs': blogs_list,
        'faqs': faqs_list,
        'services': services_list,
        'faq_submissions': faq_submissions,
        'booking_settings': booking_settings,
        'google_calendar_packages_installed': google_calendar_available(),
        'transactions': transactions_list,
        'logs': logs_list,
        'total_views': total_views,
        'scripts_this_week': scripts_this_week,
        'conversion_rate': round(conversion_rate, 1),
        'active_leads_count': active_leads_count,
        'pipeline_forecast': pipeline_forecast,
        'average_lead_score': round(average_lead_score, 1),
        'total_income': total_income,
        'total_expenses': total_expenses,
        'net_balance': net_balance,
        'salary_draw': salary_draw,
        'reinvestment_fund': reinvestment_fund,
        'funnel_counts': funnel_counts,
        'win_loss_by_project_type': win_loss_by_project_type,
        'top_pages': top_pages,
        'data_error': data_error,
        'now': datetime.utcnow(),
        'active_page': active_page,
        'selected_month': selected_month,
        'selected_year': selected_year
    }

@admin_bp.route('/admin')
@login_required
def admin():
    return render_template('admin/leads.html', **build_admin_context('leads'))


@admin_bp.route('/admin/leads')
@login_required
def leads_page():
    return render_template('admin/leads.html', **build_admin_context('leads'))


@admin_bp.route('/admin/projects')
@login_required
def projects_page():
    return render_template('admin/projects.html', **build_admin_context('projects'))


@admin_bp.route('/admin/analytics')
@login_required
def analytics_page():
    return render_template('admin/analytics.html', **build_admin_context('analytics'))


@admin_bp.route('/admin/automation')
@login_required
def automation_page():
    return render_template('admin/automation.html', **build_admin_context('automation'))


@admin_bp.route('/admin/booking')
@login_required
def booking_settings_page():
    return render_template('admin/booking.html', **build_admin_context('booking'))


@admin_bp.route('/admin/booking', methods=['POST'])
@login_required
def update_booking_settings():
    try:
        settings = get_booking_settings()
        settings.calendar_id = request.form.get('calendar_id', 'primary').strip() or 'primary'
        settings.service_account_file = request.form.get('service_account_file', '').strip()
        settings.workday_start = request.form.get('workday_start', '09:00')
        settings.workday_end = request.form.get('workday_end', '17:00')
        settings.meeting_duration_minutes = request.form.get('meeting_duration_minutes', type=int) or 30
        settings.buffer_minutes = request.form.get('buffer_minutes', type=int) or 30
        settings.slot_step_minutes = request.form.get('slot_step_minutes', type=int) or 30
        settings.booking_horizon_days = request.form.get('booking_horizon_days', type=int) or 21
        settings.min_notice_hours = request.form.get('min_notice_hours', type=int) or 4
        settings.reminder_minutes = request.form.get('reminder_minutes', type=int) or 30
        settings.create_google_meet = request.form.get('create_google_meet') == 'on'
        settings.meeting_location = request.form.get('meeting_location', 'Google Meet').strip() or 'Google Meet'

        db.session.commit()
        flash('Booking settings updated successfully.', 'success')
    except Exception as e:
        db.session.rollback()
        flash(f'Error updating booking settings: {e}', 'danger')

    return redirect(url_for('admin.booking_settings_page'))

@admin_bp.route('/admin/logout')
@login_required
def logout():
    logout_user()
    flash('You have been logged out.', 'info')
    return redirect(url_for('login'))

@admin_bp.route('/admin/analytics_data')
@login_required
def analytics_data():
    try:
        # Get views per day for the last 30 days
        thirty_days_ago = datetime.utcnow() - timedelta(days=30)
        
        # This query might vary slightly depending on the DB backend (SQLite vs Postgres)
        # For SQLite (default for many dev envs), we use strftime
        views_by_day = db.session.query(
            func.date(Analytics.timestamp).label('day'),
            func.count(Analytics.id).label('count')
        ).filter(Analytics.timestamp >= thirty_days_ago)\
         .group_by(func.date(Analytics.timestamp))\
         .order_by(func.date(Analytics.timestamp))\
         .all()
        
        labels = [v.day for v in views_by_day]
        data = [v.count for v in views_by_day]
        
        return jsonify({
            "status": "success",
            "labels": labels,
            "data": data
        })
    except Exception as e:
        return jsonify({"status": "error", "message": str(e)}), 400

@admin_bp.route('/admin/add_lead', methods=['POST'])
@login_required
def add_lead():
    try:
        data = request.get_json(silent=True)
        if not data:
            return jsonify({"status": "error", "message": "Invalid JSON"}), 400

        scoring_input = {
            "client_name": data.get('client_name') or data.get('client_company'),
            "client_company": data.get('client_company'),
            "contact_role": data.get('contact_role', 'Agent'),
            "budget": data.get('budget', 0),
            "project_type": normalize_project_type(data.get('project_type', 'static')),
            "visited_pages": data.get('visited_pages', ['/admin']),
            "whatsapp_engagement": data.get('whatsapp_engagement', 'ignored'),
            "phone_number": data.get('phone_number'),
            "last_activity_date": data.get('last_activity_date') or datetime.utcnow().strftime('%Y-%m-%d')
        }

        score_result = calculate_lead_score(scoring_input)
        
        calculated_score = score_result['score']
        lead_classification = score_result['classification']
        breakdown = score_result['breakdown']

        new_lead = Lead(
            client_name=data.get('client_name') or data.get('client_company'),
            client_company=data.get('client_company'),
            project_type=normalize_project_type(data.get('project_type')),
            budget=data.get('budget'),
            contact_role=data.get('contact_role'),
            phone_number=data.get('phone_number'),
            whatsapp_engagement=data.get('whatsapp_engagement'),
            target_project=data.get('target_project') or normalize_project_type(data.get('project_type')),
            score=int(round(calculated_score)),
            explicit_score=breakdown.get('explicit'),
            implicit_score=breakdown.get('implicit'),
            urgency_score=breakdown.get('urgency'),
            status=normalize_lead_status(data.get('status', 'New')),
            loss_reason=data.get('loss_reason'),
            last_activity_date=datetime.utcnow()
        )
        
        db.session.add(new_lead)
        db.session.commit()
        
        return jsonify({
            "status": "success",
            "message": f"Lead added. Score: {calculated_score} ({lead_classification})",
            "lead": serialize_lead(new_lead),
            "classification": lead_classification
        }), 201 

    except Exception as e:
        db.session.rollback()
        return jsonify({"status": "error", "message": str(e)}), 500

@admin_bp.route('/admin/add_project', methods=['POST'])
@login_required
def add_project():
    try:
        print("DEBUG add_project: route entered")
        title = request.form.get('project_title')
        category = request.form.get('project_tag')
        tech_stack = request.form.get('tech_stack')
        description = request.form.get('description')
        code_snippet = request.form.get('code_snippet')
        youtube_url = request.form.get('youtube_url')
        project_url = request.form.get('project_url')
        print(f"DEBUG add_project: title={title!r}, category={category!r}, tech_stack={tech_stack!r}")
        print(f"DEBUG add_project: youtube_url={youtube_url!r}, project_url={project_url!r}")
        
        # Handle file upload
        media_file = request.files.get('media')
        media_path = ""
        print(f"DEBUG add_project: media_file_present={media_file is not None}")
        if media_file and media_file.filename:
            # Simple path for now, adjust as needed
            print(f"DEBUG add_project: raw filename={media_file.filename!r}")
            filename = secure_filename(str(media_file.filename))
            print(f"DEBUG add_project: sanitized filename={filename!r}")
            save_path = os.path.join('static', 'assets', filename)
            print(f"DEBUG add_project: save_path={save_path!r}")
            os.makedirs(os.path.dirname(save_path), exist_ok=True)
            media_file.save(save_path)
            print("DEBUG add_project: file saved successfully")
            media_path = f'/static/assets/{filename}'
        else:
            print("DEBUG add_project: no media file uploaded")
            
        performance = request.form.get('performance')
        seo = request.form.get('seo')

        new_project = Project(
            title=title,
            category=category,
            tech_stack=tech_stack,
            description=description,
            code_snippet=code_snippet,
            youtube_url=youtube_url,
            project_url=project_url,
            media_path=media_path,
            performance=int(performance) if performance else None,
            seo=int(seo) if seo else None,
        )
        print(f"DEBUG add_project: new_project prepared with media_path={media_path!r}")
        db.session.add(new_project)
        print("DEBUG add_project: project added to session")
        db.session.commit()
        print("DEBUG add_project: database commit successful")
        return jsonify({"status": "success", "message": "Project deployed successfully"})
    except Exception as e:
        print(f"DEBUG add_project: exception={e!r}")
        return jsonify({"status": "error", "message": str(e)}), 400


@admin_bp.route('/admin/update_project/<int:project_id>', methods=['POST'])
@login_required
def update_project(project_id):
    try:
        title       = request.form.get('project_title', '').strip()
        category    = request.form.get('project_tag', '').strip()
        description = request.form.get('description', '').strip()
        tech_stack  = request.form.get('tech_stack', '').strip()
        youtube_url = request.form.get('youtube_url', '').strip()
        project_url = request.form.get('project_url', '').strip()
        performance_raw = request.form.get('performance', '').strip()
        seo_raw         = request.form.get('seo', '').strip()
        performance = int(performance_raw) if performance_raw else None
        seo         = int(seo_raw) if seo_raw else None

        media_path = None
        media_file = request.files.get('media')
        if media_file and media_file.filename:
            filename  = secure_filename(str(media_file.filename))
            save_path = os.path.join('static', 'assets', filename)
            os.makedirs(os.path.dirname(save_path), exist_ok=True)
            media_file.save(save_path)
            media_path = f'/static/assets/{filename}'

        if media_path:
            db.session.execute(text("""
                UPDATE projects
                SET title=:title, category=:category, description=:description,
                    tech_stack=:tech_stack, youtube_url=:youtube_url,
                    project_url=:project_url, performance=:performance,
                    seo=:seo, media_path=:media_path
                WHERE id=:id
            """), dict(title=title, category=category, description=description,
                       tech_stack=tech_stack, youtube_url=youtube_url,
                       project_url=project_url, performance=performance,
                       seo=seo, media_path=media_path, id=project_id))
        else:
            db.session.execute(text("""
                UPDATE projects
                SET title=:title, category=:category, description=:description,
                    tech_stack=:tech_stack, youtube_url=:youtube_url,
                    project_url=:project_url, performance=:performance,
                    seo=:seo
                WHERE id=:id
            """), dict(title=title, category=category, description=description,
                       tech_stack=tech_stack, youtube_url=youtube_url,
                       project_url=project_url, performance=performance,
                       seo=seo, id=project_id))

        db.session.commit()
        return jsonify({"status": "success", "message": "Project updated successfully"})
    except Exception as e:
        db.session.rollback()
        return jsonify({"status": "error", "message": str(e)}), 400


@admin_bp.route('/admin/run_script', methods=['POST'])
@login_required
def run_script():
    try:
        data = request.json
        script_name = data.get('script_name')
        
        # Here you would trigger the actual script logic
        # For now, we just log the execution
        new_log = AutomationLog(
            script_name=script_name,
            status='Success',
            timestamp=datetime.utcnow()
        )
        db.session.add(new_log)
        db.session.commit()
        
        return jsonify({"status": "success", "message": f"Script {script_name} executed successfully"})
    except Exception as e:
        return jsonify({"status": "error", "message": str(e)}), 400

@admin_bp.route('/admin/delete_lead/<int:lead_id>', methods=['POST', 'DELETE'])
@login_required
def delete_lead(lead_id):
    try:
        lead = Lead.query.get_or_404(lead_id)
        db.session.delete(lead)
        db.session.commit()
        return jsonify({"status": "success", "message": "Lead deleted successfully"})
    except Exception as e:
        return jsonify({"status": "error", "message": str(e)}), 400

@admin_bp.route('/admin/update_lead/<int:lead_id>', methods=['POST', 'PATCH'])
@login_required
def update_lead(lead_id):
    try:
        lead = Lead.query.get_or_404(lead_id)
        data = request.get_json(silent=True) or {}
        
        if 'status' in data:
            lead.status = normalize_lead_status(data['status'])
        if 'score' in data:
            lead.score = data['score']
        if 'client_name' in data:
            lead.client_name = data['client_name'] or lead.client_name
        if 'client_company' in data:
            lead.client_company = data['client_company']
        if 'project_type' in data:
            lead.project_type = normalize_project_type(data['project_type'])
            lead.target_project = normalize_project_type(data['project_type'])
        if 'budget' in data:
            lead.budget = data['budget']
        if 'contact_role' in data:
            lead.contact_role = data['contact_role']
        if 'phone_number' in data:
            lead.phone_number = data['phone_number']
        if 'whatsapp_engagement' in data:
            lead.whatsapp_engagement = data['whatsapp_engagement']
        if 'loss_reason' in data:
            lead.loss_reason = data['loss_reason']

        score_fields = {'client_name', 'client_company', 'project_type', 'budget', 'contact_role', 'phone_number', 'whatsapp_engagement', 'last_activity_date'}
        if score_fields.intersection(data.keys()) or 'status' in data:
            scoring_input = {
                "client_name": lead.client_name,
                "client_company": lead.client_company,
                "contact_role": lead.contact_role or 'Agent',
                "budget": lead.budget or 0,
                "project_type": normalize_project_type(lead.project_type or lead.target_project or 'static'),
                "visited_pages": data.get('visited_pages', ['/admin']),
                "whatsapp_engagement": lead.whatsapp_engagement or 'ignored',
                "phone_number": lead.phone_number,
                "last_activity_date": (data.get('last_activity_date') or datetime.utcnow().strftime('%Y-%m-%d'))
            }
            result = calculate_lead_score(scoring_input)
            lead.score = int(round(result['score']))
            lead.explicit_score = result['breakdown'].get('explicit')
            lead.implicit_score = result['breakdown'].get('implicit')
            lead.urgency_score = result['breakdown'].get('urgency')

        lead.last_activity_date = datetime.utcnow()
        
        db.session.commit()
        return jsonify({"status": "success", "message": "Lead updated successfully", "lead": serialize_lead(lead)})
    except Exception as e:
        return jsonify({"status": "error", "message": str(e)}), 400

@admin_bp.route('/admin/get_lead/<int:lead_id>')
@login_required
def get_lead(lead_id):
    try:
        lead = Lead.query.get_or_404(lead_id)
        return jsonify(serialize_lead(lead))
    except Exception as e:
        return jsonify({"status": "error", "message": str(e)}), 400

@admin_bp.route('/admin/blogs')
@login_required
def blogs():
    return render_template('admin/blogs.html', **build_admin_context('blogs'))


@admin_bp.route('/admin/faqs')
@login_required
def faqs():
    return render_template('admin/faqs.html', **build_admin_context('faqs'))


@admin_bp.route('/admin/services')
@login_required
def services_page():
    return render_template('admin/services.html', **build_admin_context('services'))


@admin_bp.route('/admin/add_service', methods=['POST'])
@login_required
def add_service():
    try:
        new_service = Service(
            title=request.form.get('title'),
            eyebrow=request.form.get('eyebrow'),
            lead_text=request.form.get('lead_text'),
            description=request.form.get('description'),
            features=request.form.get('features'),
            price_range=request.form.get('price_range'),
            price_label=request.form.get('price_label'),
            price_note=request.form.get('price_note'),
            icon_svg=request.form.get('icon_svg'),
            panel_title=request.form.get('panel_title'),
            panel_type=request.form.get('panel_type'),
            panel_content=request.form.get('panel_content'),
            display_order=request.form.get('display_order', type=int) or 0,
            is_published=request.form.get('is_published') == 'on'
        )
        db.session.add(new_service)
        db.session.commit()
        flash('Service added successfully!', 'success')
    except Exception as e:
        db.session.rollback()
        flash(f'Error adding service: {e}', 'danger')
    return redirect(url_for('admin.services_page'))


@admin_bp.route('/admin/update_service/<int:service_id>', methods=['POST'])
@login_required
def update_service(service_id):
    try:
        service = Service.query.get_or_404(service_id)
        service.title = request.form.get('title')
        service.eyebrow = request.form.get('eyebrow')
        service.lead_text = request.form.get('lead_text')
        service.description = request.form.get('description')
        service.features = request.form.get('features')
        service.price_range = request.form.get('price_range')
        service.price_label = request.form.get('price_label')
        service.price_note = request.form.get('price_note')
        service.icon_svg = request.form.get('icon_svg')
        service.panel_title = request.form.get('panel_title')
        service.panel_type = request.form.get('panel_type')
        service.panel_content = request.form.get('panel_content')
        service.display_order = request.form.get('display_order', type=int) or 0
        service.is_published = request.form.get('is_published') == 'on'
        
        # Explicitly update slug if title changed (slug is handled by event listener but if title is passed it might be useful)
        # But let's trust the event listener for now.
        
        db.session.commit()
        flash('Service updated successfully!', 'success')
    except Exception as e:
        db.session.rollback()
        flash(f'Error updating service: {e}', 'danger')
    return redirect(url_for('admin.services_page'))


@admin_bp.route('/admin/delete_service/<int:service_id>', methods=['POST', 'DELETE'])
@login_required
def delete_service(service_id):
    try:
        service = Service.query.get_or_404(service_id)
        db.session.delete(service)
        db.session.commit()
        return jsonify({"status": "success", "message": "Service deleted successfully"})
    except Exception as e:
        db.session.rollback()
        return jsonify({"status": "error", "message": str(e)}), 400


@admin_bp.route('/admin/delete_faq_submission/<int:sub_id>', methods=['POST', 'DELETE'])
@login_required
def delete_faq_submission(sub_id):
    try:
        sub = FAQSubmission.query.get_or_404(sub_id)
        db.session.delete(sub)
        db.session.commit()
        return jsonify({"status": "success", "message": "Submission deleted successfully"})
    except Exception as e:
        db.session.rollback()
        return jsonify({"status": "error", "message": str(e)}), 400


@admin_bp.route('/admin/add_faq', methods=['POST'])
@login_required
def add_faq():
    try:
        question = request.form.get('question', '').strip()
        answer = request.form.get('answer', '').strip()
        slug = request.form.get('slug', '').strip()
        display_order = request.form.get('display_order', type=int) or 0
        is_published = request.form.get('is_published') == 'on'

        if not question or not answer:
            flash('FAQ question and answer are required.', 'danger')
            return redirect(url_for('admin.faqs'))

        new_faq = FAQ(
            question=question,
            answer=answer,
            slug=make_slug(slug) if slug else None,
            display_order=display_order,
            is_published=is_published
        )
        db.session.add(new_faq)
        db.session.commit()
        flash('FAQ added successfully!', 'success')
    except Exception as e:
        db.session.rollback()
        flash(f'Error adding FAQ: {e}', 'danger')

    return redirect(url_for('admin.faqs'))


@admin_bp.route('/admin/update_faq/<int:faq_id>', methods=['POST'])
@login_required
def update_faq(faq_id):
    try:
        faq = FAQ.query.get_or_404(faq_id)
        question = request.form.get('question', '').strip()
        answer = request.form.get('answer', '').strip()
        slug = request.form.get('slug', '').strip()

        if not question or not answer:
            flash('FAQ question and answer are required.', 'danger')
            return redirect(url_for('admin.faqs'))

        faq.question = question
        faq.answer = answer
        faq.slug = make_slug(slug) if slug else faq.slug
        faq.display_order = request.form.get('display_order', type=int) or 0
        faq.is_published = request.form.get('is_published') == 'on'

        db.session.commit()
        flash('FAQ updated successfully!', 'success')
    except Exception as e:
        db.session.rollback()
        flash(f'Error updating FAQ: {e}', 'danger')

    return redirect(url_for('admin.faqs'))


@admin_bp.route('/admin/delete_faq/<int:faq_id>', methods=['POST', 'DELETE'])
@login_required
def delete_faq(faq_id):
    try:
        faq = FAQ.query.get_or_404(faq_id)
        db.session.delete(faq)
        db.session.commit()
        return jsonify({"status": "success", "message": "FAQ deleted successfully"})
    except Exception as e:
        db.session.rollback()
        return jsonify({"status": "error", "message": str(e)}), 400


@admin_bp.route('/admin/add_blog', methods=['POST'])
@login_required
def add_blog():
    try:
        title = request.form.get('title')
        summary = request.form.get('summary')
        content = request.form.get('content')
        
        media_path = None
        file = request.files.get('media')
        if file and file.filename:
            media_path = save_blog_media(file)

        new_post = BlogPost(
            title=title,
            summary=summary,
            content=content,
            media_path=media_path
        )
        db.session.add(new_post)
        db.session.commit()
        flash('Blog post added successfully!', 'success')
    except Exception as e:
        db.session.rollback()
        flash(f'Error adding blog post: {e}', 'danger')
        
    return redirect(url_for('admin.blogs'))


@admin_bp.route('/admin/update_blog/<int:post_id>', methods=['POST'])
@login_required
def update_blog(post_id):
    try:
        post = BlogPost.query.get_or_404(post_id)
        post.title = request.form.get('title')
        post.summary = request.form.get('summary')
        post.content = request.form.get('content')
        
        file = request.files.get('media')
        if file and file.filename:
            if post.media_path:
                old_path = os.path.join(os.getcwd(), post.media_path.lstrip('/'))
                if os.path.isfile(old_path):
                    os.remove(old_path)

            post.media_path = save_blog_media(file)

        db.session.commit()
        flash('Blog post updated successfully!', 'success')
    except Exception as e:
        db.session.rollback()
        flash(f'Error updating blog post: {e}', 'danger')
        
    return redirect(url_for('admin.blogs'))


@admin_bp.route('/admin/funds')
@login_required
def funds():
    selected_month = request.args.get('month', type=int)
    selected_year = request.args.get('year', type=int)

    context = build_admin_context('funds', selected_month=selected_month, selected_year=selected_year)
    
    # Get all unique years from transactions
    all_transaction_years = sorted(list(set([d.date.year for d in db.session.query(Transaction.date).distinct().all()])), reverse=True)
    # Get all unique months from transactions
    all_transaction_months = sorted(list(set([d.date.month for d in db.session.query(Transaction.date).distinct().all()])))
    
    context['available_years'] = all_transaction_years
    context['available_months'] = all_transaction_months
    context['selected_month'] = selected_month
    context['selected_year'] = selected_year

    return render_template('admin/funds.html', datetime=datetime, **context)


@admin_bp.route('/admin/add_transaction', methods=['POST'])
@login_required
def add_transaction():
    try:
        type = request.form.get('type')
        category = request.form.get('category')
        amount = float(request.form.get('amount', 0))
        description = request.form.get('description')
        date_str = request.form.get('date')
        
        # Parse date
        transaction_date = date.today()
        if date_str:
            transaction_date = datetime.strptime(date_str, '%Y-%m-%d').date()

        new_transaction = Transaction(
            type=type,
            category=category,
            amount=amount,
            description=description,
            date=transaction_date
        )
        db.session.add(new_transaction)
        db.session.commit()
        flash('Transaction recorded successfully!', 'success')
    except Exception as e:
        db.session.rollback()
        flash(f'Error recording transaction: {e}', 'danger')
        
    return redirect(url_for('admin.funds'))


@admin_bp.route('/admin/delete_transaction/<int:transaction_id>', methods=['POST', 'DELETE'])
@login_required
def delete_transaction(transaction_id):
    try:
        transaction = Transaction.query.get_or_404(transaction_id)
        db.session.delete(transaction)
        db.session.commit()
        return jsonify({"status": "success", "message": "Transaction deleted successfully"})
    except Exception as e:
        return jsonify({"status": "error", "message": str(e)}), 400


@admin_bp.route('/admin/export_transactions_csv')
@login_required
def export_transactions_csv():
    try:
        transactions = Transaction.query.order_by(Transaction.date.desc(), Transaction.created_at.desc()).all()
        
        si = io.StringIO()
        cw = csv.writer(si)
        
        headers = ["ID", "Type", "Category", "Amount", "Description", "Date", "Created At"]
        cw.writerow(headers)
        
        for t in transactions:
            cw.writerow([
                t.id,
                t.type,
                t.category,
                f"{t.amount:.2f}",
                t.description,
                t.date.strftime('%Y-%m-%d'),
                t.created_at.strftime('%Y-%m-%d %H:%M:%S')
            ])
            
        output = make_response(si.getvalue())
        output.headers["Content-Disposition"] = "attachment; filename=transactions_export.csv"
        output.headers["Content-type"] = "text/csv"
        return output
    except Exception as e:
        flash(f'Error exporting transactions: {e}', 'danger')
        return redirect(url_for('admin.funds'))


@admin_bp.route('/admin/delete_blog/<int:post_id>', methods=['POST', 'DELETE'])
@login_required
def delete_blog(post_id):
    try:
        post = BlogPost.query.get_or_404(post_id)
        if post.media_path:
            file_path = os.path.join(os.getcwd(), post.media_path.lstrip('/'))
            if os.path.exists(file_path):
                os.remove(file_path)
                
        db.session.delete(post)
        db.session.commit()
        return jsonify({"status": "success", "message": "Blog post deleted successfully"})
    except Exception as e:
        return jsonify({"status": "error", "message": str(e)}), 400
