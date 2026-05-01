from flask import Blueprint, render_template, abort, url_for
from database import db, Project
import json

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

@portfolio_bp.route('/projects/<slug>')
def project_detail(slug):
    project = Project.query.filter_by(slug=slug).first_or_404()

    # --- Content Strategy Reminder (Python comment block) ---
    # Golden Rule for Database Entries (Project Descriptions):
    #
    # Focus on Business Value & Metrics:
    # Instead of: "Built a Python/Flask app with a React frontend."
    # Write: "Reduced lead response time by 80% and increased conversion rates by 15% through a custom Python/Flask CRM solution."
    #
    # Target Keywords:
    # Use terms like "Custom Software Development," "CRM," "ERP," "Automation,"
    # "Business Intelligence," "Data Analytics," "Web Application Development"
    # rather than generic "Web Design" or "Website Building."
    #
    # Highlight Solutions to Client Problems:
    # Describe the challenge the client faced and how your solution specifically
    # addressed it, leading to measurable improvements.
    #
    # Examples of good descriptions:
    # - "Developed a custom ERP system for a manufacturing client, streamlining inventory management and reducing operational costs by 25%."
    # - "Implemented a secure, scalable web application for financial advisors, enabling GDPR-compliant client data management and enhancing client onboarding efficiency by 30%."
    # - "Automated critical marketing workflows for a SaaS startup, resulting in a 40% increase in qualified lead generation within six months."
    # -----------------------------------------------------------

    # Prepare JSON-LD Schema.org data for this specific project
    schema_data = {
        "@context": "https://schema.org",
        "@type": "SoftwareApplication", # or "Article" depending on content focus
        "name": project.title,
        "description": project.description,
        "applicationCategory": project.category,
        "url": url_for('portfolio.project_detail', slug=project.slug, _external=True),
        "operatingSystem": "Web", # Assuming web application
        # Add more properties as relevant, e.g., 'processorRequirements', 'memoryRequirements', 'softwareHelp', 'screenshot'
        "offers": {
            "@type": "Offer",
            "price": "Call for Quote", # Or relevant pricing info
            "priceCurrency": "USD"
        }
    }

    # Pass dynamic SEO data to the template
    return render_template(
        'projects.html',
        project=project,
        seo_title=f"{project.title} | Custom Software Development by JO4 Dev",
        seo_description=project.description, # Consider a more refined, truncated description here
        json_ld_schema=json.dumps(schema_data, indent=2)
    )

