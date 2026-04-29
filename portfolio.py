from flask import Blueprint, render_template
from database import db, Project

portfolio_bp = Blueprint('portfolio', __name__)

@portfolio_bp.route('/projects')
def projects():
    data_error = None
    try:
        projects_list = Project.query.order_by(Project.deployed_at.desc()).all()
    except Exception as e:
        db.session.rollback()
        projects_list = []
        data_error = str(e)
    return render_template('projects.html', projects=projects_list, data_error=data_error)
