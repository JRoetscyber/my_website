from flask import Flask, render_template, request, jsonify
from datetime import datetime
from werkzeug.security import generate_password_hash
from portfolio import portfolio_bp
from booking import booking_bp
from admin import admin_bp
from database import db, User, Analytics, Lead
from lead_score import calculate_lead_score

app = Flask(__name__)

# Database Configuration
app.config['SQLALCHEMY_DATABASE_URI'] = 'sqlite:///jo4dev.db'
app.config['SQLALCHEMY_TRACK_MODIFICATIONS'] = False

# Initialize DB with app
db.init_app(app)

@app.before_request
def record_visit():
    # Only record visits to static pages and main routes, ignore internal calls
    if request.endpoint and not request.path.startswith('/static') and not request.path.startswith('/admin'):
        try:
            visit = Analytics(
                page_path=request.path,
                visitor_ip=request.remote_addr
            )
            db.session.add(visit)
            db.session.commit()
        except:
            db.session.rollback()

def init_db():
    with app.app_context():
        db.create_all()
        # Insert a default 'admin' user if the login table is empty
        try:
            admin_user = User.query.filter_by(username='admin').first()
            if not admin_user:
                hashed_pw = generate_password_hash('admin123')
                new_admin = User(username='admin', password_hash=hashed_pw)
                db.session.add(new_admin)
                db.session.commit()
                print("Default admin user created.")
        except Exception as e:
            print(f"Error initializing database user: {e}")
            db.session.rollback()

# Blueprints
app.register_blueprint(portfolio_bp)
app.register_blueprint(booking_bp)
app.register_blueprint(admin_bp)

@app.context_processor
def inject_year():
    current_year = datetime.now().year
    return {'year': current_year}

@app.route('/')
def home():
    return render_template('index.html')

@app.route('/api/new-lead', methods=['POST'])
def handle_new_lead():
    """
    Automated Lead Intake & Scoring API
    Receives JSON from frontend, calculates score, and saves to DB.
    """
    try:
        data = request.json
        if not data:
            return jsonify({"status": "error", "message": "No data provided"}), 400

        # 1. Prepare data for the scoring algorithm
        # Map frontend fields to the algorithm's expected schema
        scoring_input = {
            "client_name": data.get('client_name'),
            "client_company": data.get('client_company'),
            "contact_role": data.get('contact_role', 'Agent'),
            "budget": data.get('budget', 0),
            "project_type": data.get('project_type', 'static'),
            "visited_pages": data.get('visited_pages', []),
            "whatsapp_engagement": "ignored", # New leads start as 'ignored'
            "phone_number": data.get('phone_number'),
            "last_activity_date": datetime.now().strftime('%Y-%m-%d')
        }

        # 2. Run the Senior Data Scientist's scoring algorithm
        analysis = calculate_lead_score(scoring_input)
        
        # 3. Create and Save the Lead record
        # Use the automated classification unless a specific status is provided
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

if __name__ == '__main__':
    init_db()
    app.run(debug=True, host='0.0.0.0', port=5056)
