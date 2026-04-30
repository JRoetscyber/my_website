import os
from flask import Flask, render_template, request, jsonify, redirect, url_for, flash
from datetime import datetime
from werkzeug.security import generate_password_hash, check_password_hash
from sqlalchemy import text
from flask_login import LoginManager, login_user, logout_user, login_required, current_user
from dotenv import load_dotenv

from portfolio import portfolio_bp
from booking import booking_bp
from admin import admin_bp
from database import db, User, Analytics, Lead
from lead_score import calculate_lead_score

# Load environment variables
load_dotenv()

app = Flask(__name__)

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
    """Create all tables and seed default admin user."""
    with app.app_context():
        db.create_all()

        # Migrate projects table if needed
        try:
            project_cols = {
                row[1] for row in db.session.execute(text("PRAGMA table_info(projects)")).fetchall()
            }
            if 'youtube_url' not in project_cols:
                db.session.execute(text("ALTER TABLE projects ADD COLUMN youtube_url TEXT"))
            if 'project_url' not in project_cols:
                db.session.execute(text("ALTER TABLE projects ADD COLUMN project_url TEXT"))
            db.session.commit()
        except Exception as e:
            print(f"Error migrating projects table: {e}")
            db.session.rollback()

        # Seed default admin user if none exists
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

# Blueprints
app.register_blueprint(portfolio_bp)
app.register_blueprint(booking_bp)
app.register_blueprint(admin_bp)

@app.context_processor
def inject_year():
    return {'year': datetime.now().year}

@app.route('/')
def home():
    return render_template('index.html')

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

        manual_status = data.get('status')
        final_status = manual_status if manual_status and manual_status != 'New' else analysis['classification']

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


# Initialise the database when the module is imported (by gunicorn/wsgi)
init_db()

if __name__ == '__main__':
    app.run(debug=True, host='0.0.0.0', port=6010)