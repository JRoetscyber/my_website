from flask import Blueprint, render_template, abort
from database import db, Service, FAQ

services_bp = Blueprint('services', __name__)

@services_bp.route('/services')
def services():
    services_list = Service.query.filter_by(is_published=True).order_by(Service.display_order.asc()).all()
    # Fetch top 6 FAQs for the bottom section
    faqs = FAQ.query.filter_by(is_published=True).order_by(FAQ.display_order.asc()).limit(6).all()
    return render_template('services.html', services=services_list, faqs=faqs)

@services_bp.route('/services/<slug>')
def service_detail(slug):
    service = Service.query.filter_by(slug=slug, is_published=True).first_or_404()
    return render_template('service_detail.html', service=service)
