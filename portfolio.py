from flask import Blueprint, render_template

portfolio_bp = Blueprint('portfolio', __name__)

@portfolio_bp.route('/projects')
def projects():
    return render_template('projects.html')