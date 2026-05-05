from flask import Blueprint, render_template, request, redirect, url_for, flash, jsonify
from flask_login import login_required, logout_user
from database import db, Lead, Project, AutomationLog, Analytics
from datetime import datetime, timedelta
from sqlalchemy import func
from lead_score import calculate_lead_score
import os
import re
import unicodedata

def secure_filename(filename):
    filename = str(filename)
    filename = unicodedata.normalize('NFKD', filename).encode('ascii', 'ignore').decode('ascii')
    filename = filename.replace(os.path.sep, ' ')
    if os.path.altsep:
        filename = filename.replace(os.path.altsep, ' ')
    filename = re.sub(r'[^A-Za-z0-9_.-]+', '_', filename).strip('._')
    return filename or 'upload'


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


def build_admin_context(active_page):
    data_error = None
    top_pages = []

    try:
        leads_list = Lead.query.order_by(Lead.created_at.desc(), Lead.id.desc()).all()
        leads_list = prepare_leads_for_display(leads_list)
        projects_list = Project.query.order_by(Project.deployed_at.desc()).all()
        logs_list = AutomationLog.query.order_by(AutomationLog.timestamp.desc()).limit(10).all()

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
        logs_list = []
        total_views = 0
        scripts_this_week = 0
        conversion_rate = 0
        active_leads_count = 0
        pipeline_forecast = 0
        average_lead_score = 0
        funnel_counts = {'new': 0, 'contacted': 0, 'negotiating': 0, 'closed': 0}
        win_loss_by_project_type = []

    return {
        'leads': leads_list,
        'projects': projects_list,
        'logs': logs_list,
        'total_views': total_views,
        'scripts_this_week': scripts_this_week,
        'conversion_rate': round(conversion_rate, 1),
        'active_leads_count': active_leads_count,
        'pipeline_forecast': pipeline_forecast,
        'average_lead_score': round(average_lead_score, 1),
        'funnel_counts': funnel_counts,
        'win_loss_by_project_type': win_loss_by_project_type,
        'top_pages': top_pages,
        'data_error': data_error,
        'now': datetime.utcnow(),
        'active_page': active_page
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
            
        new_project = Project(
            title=title,
            category=category,
            tech_stack=tech_stack,
            description=description,
            code_snippet=code_snippet,
            youtube_url=youtube_url,
            project_url=project_url,
            media_path=media_path
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

@admin_bp.route('/admin/delete_project/<int:project_id>', methods=['POST', 'DELETE'])
@login_required
def delete_project(project_id):
    try:
        project = Project.query.get_or_404(project_id)
        # Optionally delete the file from storage
        if project.media_path:
            file_path = os.path.join(os.getcwd(), project.media_path.lstrip('/'))
            if os.path.exists(file_path):
                os.remove(file_path)
                
        db.session.delete(project)
        db.session.commit()
        return jsonify({"status": "success", "message": "Project deleted successfully"})
    except Exception as e:
        return jsonify({"status": "error", "message": str(e)}), 400
