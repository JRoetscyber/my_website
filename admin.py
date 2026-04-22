from flask import Blueprint, render_template, request, redirect, url_for, flash, jsonify
from database import db, Lead, Project, AutomationLog, Analytics
from datetime import datetime, timedelta
from sqlalchemy import func
import os

admin_bp = Blueprint('admin', __name__)

@admin_bp.route('/admin')
def admin():
    # Query all data for the admin sections
    leads_list = Lead.query.order_by(Lead.created_at.desc()).all()
    projects_list = Project.query.order_by(Project.deployed_at.desc()).all()
    logs_list = AutomationLog.query.order_by(AutomationLog.timestamp.desc()).limit(10).all()
    
    # --- ANALYTICS CALCULATIONS ---
    # 1. Total Portfolio Views (from Analytics table)
    total_views = Analytics.query.count()
    
    # 2. Scripts Executed This Week
    one_week_ago = datetime.utcnow() - timedelta(days=7)
    scripts_this_week = AutomationLog.query.filter(AutomationLog.timestamp >= one_week_ago).count()
    
    # 3. Lead Conversion Rate
    total_leads = len(leads_list)
    closed_leads = Lead.query.filter_by(status='Closed').count()
    conversion_rate = (closed_leads / total_leads * 100) if total_leads > 0 else 0
    
    # 4. Active Leads (total in pipeline)
    active_leads_count = total_leads
    
    # 5. Top Pages (Dynamic)
    top_pages_query = db.session.query(
        Analytics.page_path, 
        func.count(Analytics.id).label('count')
    ).group_by(Analytics.page_path).order_by(func.count(Analytics.id).desc()).limit(5).all()
    
    # Convert to list of dicts for template
    top_pages = []
    max_count = top_pages_query[0].count if top_pages_query else 1
    for p in top_pages_query:
        top_pages.append({
            'path': p.page_path,
            'count': p.count,
            'percentage': (p.count / max_count * 100) if max_count > 0 else 0
        })
    
    return render_template('admin.html', 
                           leads=leads_list, 
                           projects=projects_list, 
                           logs=logs_list,
                           total_views=total_views,
                           scripts_this_week=scripts_this_week,
                           conversion_rate=round(conversion_rate, 1),
                           active_leads_count=active_leads_count,
                           top_pages=top_pages)

@admin_bp.route('/admin/analytics_data')
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
def add_lead():
    try:
        data = request.json
        new_lead = Lead(
            client_name=data.get('client_name'),
            client_company=data.get('client_company'),
            target_project=data.get('target_project'),
            score=data.get('score', 50),
            status=data.get('status', 'New')
        )
        db.session.add(new_lead)
        db.session.commit()
        return jsonify({"status": "success", "message": "Lead added successfully"})
    except Exception as e:
        return jsonify({"status": "error", "message": str(e)}), 400

@admin_bp.route('/admin/add_project', methods=['POST'])
def add_project():
    try:
        title = request.form.get('project_title')
        category = request.form.get('project_tag')
        tech_stack = request.form.get('tech_stack')
        description = request.form.get('description')
        code_snippet = request.form.get('code_snippet')
        
        # Handle file upload
        media_file = request.files.get('media')
        media_path = ""
        if media_file:
            # Simple path for now, adjust as needed
            filename = media_file.filename
            save_path = os.path.join('static', 'assets', filename)
            media_file.save(save_path)
            media_path = f'/static/assets/{filename}'
            
        new_project = Project(
            title=title,
            category=category,
            tech_stack=tech_stack,
            description=description,
            code_snippet=code_snippet,
            media_path=media_path
        )
        db.session.add(new_project)
        db.session.commit()
        return jsonify({"status": "success", "message": "Project deployed successfully"})
    except Exception as e:
        return jsonify({"status": "error", "message": str(e)}), 400

@admin_bp.route('/admin/run_script', methods=['POST'])
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
def delete_lead(lead_id):
    try:
        lead = Lead.query.get_or_404(lead_id)
        db.session.delete(lead)
        db.session.commit()
        return jsonify({"status": "success", "message": "Lead deleted successfully"})
    except Exception as e:
        return jsonify({"status": "error", "message": str(e)}), 400

@admin_bp.route('/admin/update_lead/<int:lead_id>', methods=['POST', 'PATCH'])
def update_lead(lead_id):
    try:
        lead = Lead.query.get_or_404(lead_id)
        data = request.json
        
        if 'status' in data:
            lead.status = data['status']
        if 'score' in data:
            lead.score = data['score']
        
        db.session.commit()
        return jsonify({"status": "success", "message": "Lead updated successfully"})
    except Exception as e:
        return jsonify({"status": "error", "message": str(e)}), 400

@admin_bp.route('/admin/get_lead/<int:lead_id>')
def get_lead(lead_id):
    try:
        lead = Lead.query.get_or_404(lead_id)
        return jsonify({
            "id": lead.id,
            "client_name": lead.client_name,
            "client_company": lead.client_company,
            "target_project": lead.target_project,
            "score": lead.score,
            "status": lead.status
        })
    except Exception as e:
        return jsonify({"status": "error", "message": str(e)}), 400

@admin_bp.route('/admin/delete_project/<int:project_id>', methods=['POST', 'DELETE'])
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
