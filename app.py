import os
import click
import markdown
from flask import Flask, render_template, request, jsonify, redirect, url_for, flash, Response, send_from_directory, current_app
from datetime import datetime
from werkzeug.security import generate_password_hash, check_password_hash
from sqlalchemy import text
from flask_login import LoginManager, login_user, logout_user, login_required, current_user
from dotenv import load_dotenv
from lead_score import calculate_lead_score

from portfolio import portfolio_bp
from booking import booking_bp
from admin import admin_bp
from blog import blog_bp
from services import services_bp
from database import db, User, Analytics, Lead, Project, BlogPost, FAQ, Transaction, Service, make_slug # Added Project, BlogPost, Transaction
from legacy_leads import backfill_legacy_leads
from faq import faq_bp

# Load environment variables
load_dotenv()

app = Flask(__name__)

def versioned_static(filename):
    static_path = os.path.join(app.static_folder, filename)
    version = None
    try:
        version = int(os.path.getmtime(static_path))
    except OSError:
        pass

    if version:
        return url_for('static', filename=filename, v=version)
    return url_for('static', filename=filename)


def versioned_asset(path):
    if not path:
        return path

    static_prefix = '/static/'
    if path.startswith(static_prefix):
        return versioned_static(path[len(static_prefix):])

    return path


@app.context_processor
def inject_asset_helpers():
    return {
        'static_url': versioned_static,
        'asset_url': versioned_asset
    }

@app.template_filter('markdown')
def markdown_filter(text):
    if not text:
        return ""
    return markdown.markdown(text, extensions=['fenced_code', 'tables', 'nl2br'])

@app.template_filter('from_json')
def from_json_filter(value):
    import json
    if not value:
        return []
    try:
        return json.loads(value)
    except Exception:
        return []

# Configuration
app.config['SECRET_KEY'] = os.getenv('SECRET_KEY', 'dev-key-123')

# SQLite DB will be written to /app/instance/jo4dev.db inside the container.
# The instance folder is mounted as a Docker named volume so it persists.
app.config['SQLALCHEMY_DATABASE_URI'] = os.getenv(
    'SQLALCHEMY_DATABASE_URI',
    'sqlite:////app/instance/jo4dev.db'   # <-- 4 slashes = absolute path
)
app.config['SQLALCHEMY_TRACK_MODIFICATIONS'] = False

# Initialize DB with app
db.init_app(app)

# Initialize Flask-Login
login_manager = LoginManager()
login_manager.init_app(app)
login_manager.login_view = 'login'

@login_manager.user_loader
def load_user(user_id):
    return User.query.get(int(user_id))

# --------------------------------------------------
# NEW: Flask-Mail & Brevo Setup
# --------------------------------------------------
from flask_mail import Mail

app.config['MAIL_SERVER'] = os.getenv('MAIL_SERVER')
app.config['MAIL_PORT'] = int(os.getenv('MAIL_PORT', 587))
app.config['MAIL_USE_TLS'] = True
app.config['MAIL_USERNAME'] = os.getenv('MAIL_USERNAME')
app.config['MAIL_PASSWORD'] = os.getenv('MAIL_PASSWORD')
app.config['MAIL_DEFAULT_SENDER'] = os.getenv('MAIL_DEFAULT_SENDER')

mail = Mail(app)
# --------------------------------------------------

@app.before_request
def record_visit():
    if request.endpoint and not request.path.startswith('/static') and not request.path.startswith('/admin'):
        try:
            visit = Analytics(
                page_path=request.path,
                visitor_ip=request.remote_addr
            )
            db.session.add(visit)
            db.session.commit()
        except Exception as e:
            db.session.rollback()

def init_db():
    """Create all tables, migrate schema, and ensure all projects have slugs."""
    with app.app_context():
        # Ensure the instance folder exists for the SQLite database
        os.makedirs('/app/instance', exist_ok=True) 
        
        db.create_all()

        # Migrate leads table schema
        try:
            lead_cols = {
                row[1] for row in db.session.execute(text("PRAGMA table_info(leads)")).fetchall()
            }

            if 'project_type' not in lead_cols:
                db.session.execute(text("ALTER TABLE leads ADD COLUMN project_type VARCHAR(100)"))
            if 'budget' not in lead_cols:
                db.session.execute(text("ALTER TABLE leads ADD COLUMN budget FLOAT"))
            if 'contact_role' not in lead_cols:
                db.session.execute(text("ALTER TABLE leads ADD COLUMN contact_role VARCHAR(100)"))
            if 'phone_number' not in lead_cols:
                db.session.execute(text("ALTER TABLE leads ADD COLUMN phone_number VARCHAR(50)"))
            if 'whatsapp_engagement' not in lead_cols:
                db.session.execute(text("ALTER TABLE leads ADD COLUMN whatsapp_engagement VARCHAR(50)"))
            if 'explicit_score' not in lead_cols:
                db.session.execute(text("ALTER TABLE leads ADD COLUMN explicit_score FLOAT"))
            if 'implicit_score' not in lead_cols:
                db.session.execute(text("ALTER TABLE leads ADD COLUMN implicit_score FLOAT"))
            if 'urgency_score' not in lead_cols:
                db.session.execute(text("ALTER TABLE leads ADD COLUMN urgency_score FLOAT"))
            if 'loss_reason' not in lead_cols:
                db.session.execute(text("ALTER TABLE leads ADD COLUMN loss_reason VARCHAR(255)"))
            if 'last_activity_date' not in lead_cols:
                db.session.execute(text("ALTER TABLE leads ADD COLUMN last_activity_date DATETIME"))

            db.session.commit()
        except Exception as e:
            print(f"Error migrating leads table: {e}")
            db.session.rollback()

        # Migrate services table schema
        try:
            service_cols = {
                row[1] for row in db.session.execute(text("PRAGMA table_info(services)")).fetchall()
            }
            if 'has_dedicated_page' not in service_cols:
                db.session.execute(text("ALTER TABLE services ADD COLUMN has_dedicated_page BOOLEAN DEFAULT 0"))
            db.session.commit()
        except Exception as e:
            print(f"Error migrating services table: {e}")
            db.session.rollback()

        # Migrate projects table schema
        try:
            project_cols = {
                row[1] for row in db.session.execute(text("PRAGMA table_info(projects)")).fetchall()
            }
            
            # Add missing columns if they don't exist
            if 'youtube_url' not in project_cols:
                db.session.execute(text("ALTER TABLE projects ADD COLUMN youtube_url TEXT"))
            if 'project_url' not in project_cols:
                db.session.execute(text("ALTER TABLE projects ADD COLUMN project_url TEXT"))
            if 'slug' not in project_cols:
                db.session.execute(text("ALTER TABLE projects ADD COLUMN slug VARCHAR(200)"))
            if 'performance' not in project_cols:
                db.session.execute(text("ALTER TABLE projects ADD COLUMN performance INTEGER"))
            if 'seo' not in project_cols:
                db.session.execute(text("ALTER TABLE projects ADD COLUMN seo INTEGER"))
            
            db.session.commit()

            # Fix data: Generate missing slugs for existing projects[cite: 3]
            # This prevents sitemap BuildErrors caused by empty slug values[cite: 3]
            projects_to_fix = Project.query.filter((Project.slug == None) | (Project.slug == '')).all()
            if projects_to_fix:
                print(f"Repairing {len(projects_to_fix)} project slugs...")
                for project in projects_to_fix:
                    if project.title:
                        project.slug = make_slug(project.title)
                db.session.commit()

        except Exception as e:
            print(f"Error migrating projects table: {e}")
            db.session.rollback()

        # Seed default admin user if none exists[cite: 3]
        try:
            admin_username = os.getenv('ADMIN_USERNAME', 'admin')
            admin_password = os.getenv('ADMIN_PASSWORD', 'admin123')
            admin_user = User.query.filter_by(username=admin_username).first()
            if not admin_user:
                hashed_pw = generate_password_hash(admin_password)
                new_admin = User(username=admin_username, password_hash=hashed_pw)
                db.session.add(new_admin)
                db.session.commit()
                print(f"Default admin user '{admin_username}' created.")
        except Exception as e:
            print(f"Error initializing database user: {e}")
            db.session.rollback()

        try:
            faq_cols = {
                row[1] for row in db.session.execute(text("PRAGMA table_info(faqs)")).fetchall()
            }

            if faq_cols:
                if 'slug' not in faq_cols:
                    db.session.execute(text("ALTER TABLE faqs ADD COLUMN slug VARCHAR(255)"))
                if 'display_order' not in faq_cols:
                    db.session.execute(text("ALTER TABLE faqs ADD COLUMN display_order INTEGER DEFAULT 0"))
                if 'is_published' not in faq_cols:
                    db.session.execute(text("ALTER TABLE faqs ADD COLUMN is_published BOOLEAN DEFAULT 1"))
                if 'created_at' not in faq_cols:
                    db.session.execute(text("ALTER TABLE faqs ADD COLUMN created_at DATETIME"))
                if 'updated_at' not in faq_cols:
                    db.session.execute(text("ALTER TABLE faqs ADD COLUMN updated_at DATETIME"))
                db.session.commit()

                faqs_to_fix = FAQ.query.filter((FAQ.slug == None) | (FAQ.slug == '')).all()
                for faq in faqs_to_fix:
                    faq.slug = make_slug(faq.question)
                if faqs_to_fix:
                    db.session.commit()

                if FAQ.query.count() == 0:
                    default_faqs = [
                        FAQ(
                            question="What kind of businesses do you build systems for?",
                            answer="JO4 Dev builds custom web portals, dashboards, CRM pipelines, booking flows, and automation systems for service businesses, property teams, agencies, and growing companies that need better internal tools.",
                            display_order=10,
                            is_published=True
                        ),
                        FAQ(
                            question="Do I own the website or system after delivery?",
                            answer="Yes. You own the delivered source code and can host it where you choose. JO4 Dev can also manage deployment, SSL, backups, and monitoring if you want ongoing support.",
                            display_order=20,
                            is_published=True
                        ),
                        FAQ(
                            question="Can you improve an existing website instead of rebuilding it?",
                            answer="Yes. Existing sites can be audited, cleaned up, secured, optimized, or extended with new portals, integrations, analytics, lead capture, and automation features.",
                            display_order=30,
                            is_published=True
                        ),
                        FAQ(
                            question="How does a project usually start?",
                            answer="Most projects start with a free discovery call to understand the business workflow, current tools, must-have features, timeline, and budget range before a practical build plan is proposed.",
                            display_order=40,
                            is_published=True
                        ),
                        FAQ(
                            question="Do you offer hosting and maintenance?",
                            answer="Yes. JO4 Dev can provide managed VPS hosting, SSL setup, database backups, uptime monitoring, and ongoing improvements after launch.",
                            display_order=50,
                            is_published=True
                        )
                    ]
                    db.session.add_all(default_faqs)
                    db.session.commit()
        except Exception as e:
            print(f"Error migrating FAQ table: {e}")
            db.session.rollback()

        # Migrate services table schema
        try:
            service_cols = {
                row[1] for row in db.session.execute(text("PRAGMA table_info(services)")).fetchall()
            }
            
            if service_cols:
                # Add missing columns if they don't exist (if any were added later)
                pass
            
            db.session.commit()

            if Service.query.count() == 0:
                import json
                default_services = [
                    Service(
                        title="Web Design & Development",
                        slug="web-design",
                        eyebrow="Service 01",
                        lead_text="Your website is your best salesperson. For construction companies and contractors across South Africa, a slow, outdated site isn't just embarrassing — it's costing you quotes every single day.",
                        description="We build bespoke websites engineered to convert — fast-loading, mobile-first, and structured so Google knows exactly what you do and where you do it. No page builders, no templates someone else already owns, no monthly platform fees to keep your own site live.",
                        features="Custom design — no templates, no drag-and-drop builders\nMobile-first, sub-3-second load times\nOn-page SEO baked in from the first line of code\nLead capture forms wired to your CRM or email\nFull source code ownership — you take it with you\nProject gallery & testimonial sections built for conversion",
                        price_range="R15 000 – R50 000",
                        price_label="Typical project range",
                        price_note="Scoped to your requirements — no surprise invoices",
                        icon_svg='<polygon points="12 2 22 8.5 22 15.5 12 22 2 15.5 2 8.5"/><line x1="12" y1="22" x2="12" y2="15.5"/><polyline points="22 8.5 12 15.5 2 8.5"/>',
                        panel_title="Who this is for",
                        panel_type="use-case",
                        panel_content=json.dumps([
                            {"title": "Construction & contractors", "desc": "Showcase your projects, get quote requests online, and stop losing work to competitors with better-looking sites.", "icon": "monitor"},
                            {"title": "Real estate agencies", "desc": "Property portals, listing management, and lead capture systems that work while you sleep.", "icon": "home"},
                            {"title": "E-commerce & retail", "desc": "Custom online stores built for South African payment gateways and shipping workflows.", "icon": "shopping-cart"}
                        ]),
                        display_order=10
                    ),
                    Service(
                        title="Local SEO & Technical Audits",
                        slug="seo",
                        eyebrow="Service 02",
                        lead_text="A construction company in Johannesburg that doesn't appear on the first page of Google for 'contractor near me' is invisible. SEO fixes that — but only when it's done properly, not as an afterthought.",
                        description="We start every engagement with a full technical audit so you know exactly what's broken and what it'll take to rank. No vague monthly retainers with nothing to show for them. Every deliverable is documented, every change is explained, and you own the results.",
                        features="Full technical SEO audit with prioritised action plan\nLocal keyword research targeting your city & service area\nOn-page optimisation — titles, meta, headings, schema markup\nGoogle Search Console & Analytics setup & monitoring\nCompetitor gap report — find where you're losing and why",
                        price_range="R15 000 – R35 000",
                        price_label="Typical project range",
                        price_note="Audit + implementation or audit-only available",
                        icon_svg='<circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/><path d="M11 8v6M8 11h6"/>',
                        panel_title="What a SEO audit covers",
                        panel_type="audit",
                        panel_content=json.dumps([
                            {"num": "01", "title": "Technical health check", "desc": "Crawl errors, broken links, page speed, Core Web Vitals, mobile usability."},
                            {"num": "02", "title": "On-page structure", "desc": "Title tags, meta descriptions, heading hierarchy, keyword mapping per page."},
                            {"num": "03", "title": "Local signal strength", "desc": "NAP consistency, local keyword coverage, GBP alignment, citation gaps."},
                            {"num": "04", "title": "Competitor gap analysis", "desc": "Where your competitors rank that you don't — and exactly what to do about it."}
                        ]),
                        display_order=20
                    ),
                    Service(
                        title="Google Business Profile Setup & Optimisation",
                        slug="gbp",
                        eyebrow="Service 03",
                        lead_text="When a homeowner in Pretoria searches 'building contractor near me', the businesses in the Map Pack get the call. That three-pack is owned by whoever has the best-optimised Google Business Profile.",
                        description="Most contractors have either a missing, unclaimed, or badly filled-in GBP. We set it up properly from scratch or audit and fix what's there — categories, services, photos, Q&A, review strategy and all — so you appear where your customers are actually looking.",
                        features="Full profile setup or audit & clean-up\nPrimary & secondary category selection (most get this wrong)\nServices, products & attributes fully populated\nKeyword-optimised business description\nPhoto & post strategy to signal active business\nReview acquisition system so 5-stars build automatically",
                        price_range="R5 000 – R15 000",
                        price_label="Typical project range",
                        price_note="One-time setup or ongoing management available",
                        icon_svg='<path d="M21 10c0 7-9 13-9 13S3 17 3 10a9 9 0 1 1 18 0z"/><circle cx="12" cy="10" r="3"/>',
                        panel_title="The Map Pack opportunity",
                        panel_type="stats",
                        panel_content=json.dumps([
                            {"num": "46", "suffix": "%", "desc": "of all Google searches have local intent — people looking for something near them"},
                            {"num": "76", "suffix": "%", "desc": "of people who search for a local business on mobile visit within 24 hours"},
                            {"num": "3", "suffix": "", "desc": "spots in the Google Map Pack — most of your competitors aren't in it yet"}
                        ]),
                        display_order=30
                    ),
                    Service(
                        title="Software & Business Automation",
                        slug="automation",
                        eyebrow="Service 04",
                        lead_text="Most construction businesses are running on WhatsApp groups, scattered spreadsheets, and memory. Every hour your team spends chasing follow-ups or copying data is an hour not on the tools.",
                        description="We build custom automation systems — CRM pipelines, quote workflows, lead scoring, and client portals — that plug into the tools you already use and eliminate the manual work that eats your margin. Built on open-source tech you own, not a SaaS subscription you're locked into.",
                        features="Custom CRM built around your actual sales process\nWhatsApp & email automation via your existing number\nQuote, invoice & job-tracking workflows\nClient-facing portal — project status, documents, sign-off\nIntegrations: Xero, QuickBooks, Google Sheets, and more",
                        price_range="R20 000 – R50 000",
                        price_label="Typical project range",
                        price_note="Scoped after a free discovery call",
                        icon_svg='<polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/>',
                        panel_title="What gets automated",
                        panel_type="auto",
                        panel_content=json.dumps([
                            {"title": "Lead follow-up", "desc": "New quote request → instant WhatsApp or email acknowledgement → CRM entry. No leads fall through."},
                            {"title": "Quote & invoice workflows", "desc": "Generate, send and track quotes from one system — not scattered across WhatsApp and email threads."},
                            {"title": "Project status updates", "desc": "Clients get automatic progress notifications — fewer 'just checking in' calls, more trust."},
                            {"title": "Review requests", "desc": "Job completed → automated request for a Google review while the client is still happy."}
                        ]),
                        display_order=40
                    )
                ]
                db.session.add_all(default_services)
                db.session.commit()
        except Exception as e:
            print(f"Error migrating Service table: {e}")
            db.session.rollback()
            
# Blueprints
app.register_blueprint(portfolio_bp)
app.register_blueprint(booking_bp)
app.register_blueprint(admin_bp)
app.register_blueprint(blog_bp)
app.register_blueprint(faq_bp)
app.register_blueprint(services_bp)
print("DEBUG: Blog blueprint registered.")

@app.context_processor
def inject_year():
    return {'year': datetime.now().year}

@app.route('/')
def home():
    projects = Project.query.filter(Project.slug != None, Project.slug != '').order_by(Project.deployed_at.desc()).limit(3).all()
    services = Service.query.filter_by(is_published=True).order_by(Service.display_order.asc()).all()
    # Also fetch recent blog posts for the home page if desired, but for now let's keep it simple
    return render_template('index.html', projects=projects, services=services)

@app.route('/login', methods=['GET', 'POST'])
def login():
    if current_user.is_authenticated:
        return redirect(url_for('admin.admin'))

    if request.method == 'POST':
        username = request.form.get('username')
        password = request.form.get('password')
        user = User.query.filter_by(username=username).first()

        if user and check_password_hash(user.password_hash, password):
            login_user(user)
            next_page = request.args.get('next')
            return redirect(next_page or url_for('admin.admin'))
        else:
            flash('Invalid username or password', 'danger')

    return render_template('login.html')

@app.route('/favicon.ico')
def favicon():
    return send_from_directory(current_app.root_path, 'favicon.ico', mimetype='image/vnd.microsoft.icon')

@app.route('/robots.txt')
def robots_txt():
    """Serve the robots.txt file from the root directory."""
    return send_from_directory(current_app.root_path, 'robots.txt', mimetype='text/plain')

@app.route('/sitemap')
@app.route('/sitemap.xml')
def sitemap_xml():
    """Generate and serve the sitemap.xml file."""
    projects = Project.query.all()
    blog_posts = BlogPost.query.all()
    faqs = FAQ.query.filter_by(is_published=True).all()
    services = Service.query.filter_by(is_published=True).all()
    
    # You can add other static URLs here if needed
    static_urls = [
        url_for('home', _external=True, _scheme='https'),
        url_for('portfolio.projects', _external=True, _scheme='https'),
        url_for('blog.blog_list', _external=True, _scheme='https'),
        url_for('faq.faq_list', _external=True, _scheme='https'),
        url_for('booking.book', _external=True, _scheme='https'),
        url_for('services.services', _external=True, _scheme='https')
    ]
    
    # Passing now=datetime.now() prevents the HTTP 500 crash
    return render_template(
        'sitemap.xml', 
        projects=projects, 
        blog_posts=blog_posts,
        faqs=faqs,
        services=services,
        static_urls=static_urls,
        now=datetime.now()
    ), 200, {'Content-Type': 'application/xml'}


@app.errorhandler(404)
def page_not_found(e):
    """Custom 404 error handler."""
    return render_template('index.html'), 404


@app.route('/api/new-lead', methods=['POST'])
def handle_new_lead():
    """Automated Lead Intake & Scoring API."""
    try:
        data = request.json
        if not data:
            return jsonify({"status": "error", "message": "No data provided"}), 400

        scoring_input = {
            "client_name": data.get('client_name'),
            "client_company": data.get('client_company'),
            "contact_role": data.get('contact_role', 'Agent'),
            "budget": data.get('budget', 0),
            "project_type": data.get('project_type', 'static'),
            "visited_pages": data.get('visited_pages', []),
            "whatsapp_engagement": "ignored",
            "phone_number": data.get('phone_number'),
            "last_activity_date": datetime.now().strftime('%Y-%m-%d')
        }

        analysis = calculate_lead_score(scoring_input)

        valid_statuses = {'New', 'Contacted', 'Negotiating', 'Closed', 'Lost'}
        manual_status = data.get('status')
        final_status = manual_status if manual_status in valid_statuses else 'New'

        new_lead = Lead(
            client_name=data.get('client_name'),
            client_company=data.get('client_company'),
            target_project=data.get('project_type'),
            score=int(analysis['score']),
            status=final_status
        )

        db.session.add(new_lead)
        db.session.commit()

        return jsonify({
            "status": "success",
            "message": "Lead captured and scored successfully",
            "lead_id": new_lead.id,
            "automated_score": analysis['score'],
            "classification": analysis['classification']
        }), 201

    except Exception as e:
        db.session.rollback()
        return jsonify({"status": "error", "message": str(e)}), 500


@app.cli.command("backfill-legacy-leads")
def backfill_legacy_leads_command():
    """Populate missing lead scoring/status data for legacy CRM rows."""
    updated = backfill_legacy_leads(app, verbose=False)
    click.echo(f"Backfilled {updated} legacy leads.")


# Initialise the database when the module is imported (by gunicorn/wsgi)
@app.cli.command("init-db")
def init_db_command():
    """Clear the existing data and create new tables."""
    init_db()
    print("Database initialized and migrated.")

if __name__ == '__main__':
    # Safe to run locally, but ignored by Gunicorn in production
    init_db()
    app.run(debug=True, host='0.0.0.0', port=6010)
