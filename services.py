from flask import Blueprint, render_template
from database import db, Service, FAQ

services_bp = Blueprint('services', __name__)

@services_bp.route('/services')
def services():
    services_list = Service.query.filter_by(is_published=True).order_by(Service.display_order.asc()).all()
    faqs = FAQ.query.filter_by(is_published=True).order_by(FAQ.display_order.asc()).limit(6).all()
    return render_template('services.html', services=services_list, faqs=faqs)

@services_bp.route('/web-design-services') # Optimized to match "web-design-services" (SV: 500)
def web_design():
    return render_template('web_design.html')

@services_bp.route('/seo-services') # Optimized to match "seo-services" (SV: 1.3K)
def seo_services():
    return render_template('seo.html')

@services_bp.route('/application-security-testing') # Optimized to match "application security testing"
def appsec():
    return render_template('appsec.html')

@services_bp.route('/business-automation-software') # Optimized to match "business automation software" (SV: 1.6K)
def automation():
    return render_template('automation.html')