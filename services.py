from flask import Blueprint, render_template, abort
from database import db, Service

services_bp = Blueprint('services', __name__)

@services_bp.route('/services')
def services():
    services_list = Service.query.filter_by(is_published=True).order_by(Service.display_order.asc()).all()
    return render_template('services.html', services=services_list)

@services_bp.route('/services/<slug>')
def service_detail(slug):
    service = Service.query.filter_by(slug=slug, is_published=True).first_or_404()
    return render_template('service_detail.html', service=service)
