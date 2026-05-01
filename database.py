from flask_sqlalchemy import SQLAlchemy
from datetime import datetime
from flask_login import UserMixin
from sqlalchemy import event
from slugify import slugify # Ensure you have 'python-slugify' installed: pip install python-slugify

db = SQLAlchemy()

class Lead(db.Model):
    __tablename__ = 'leads'
    id = db.Column(db.Integer, primary_key=True)
    client_name = db.Column(db.String(100), nullable=False)
    client_company = db.Column(db.String(100))
    target_project = db.Column(db.String(100))
    score = db.Column(db.Integer)
    status = db.Column(db.String(50), default='New') # 'New', 'Contacted', 'Proposal Sent', 'Closed'
    created_at = db.Column(db.DateTime, default=datetime.utcnow)

class Project(db.Model):
    __tablename__ = 'projects'
    id = db.Column(db.Integer, primary_key=True)
    title = db.Column(db.String(100), nullable=False)
    slug = db.Column(db.String(200), unique=True, nullable=False, index=True) # Added slug field
    category = db.Column(db.String(50))
    tech_stack = db.Column(db.Text) # Stored as comma-separated string
    description = db.Column(db.Text)
    code_snippet = db.Column(db.Text)
    youtube_url = db.Column(db.String(500))
    project_url = db.Column(db.String(500))
    media_path = db.Column(db.String(255))
    views = db.Column(db.Integer, default=0)
    deployed_at = db.Column(db.DateTime, default=datetime.utcnow)

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        if not self.slug:
            self.slug = self._generate_unique_slug(self.title)

    def _generate_unique_slug(self, title):
        if not title:
            return ""
        base_slug = slugify(title)
        slug = base_slug
        counter = 1
        while Project.query.filter_by(slug=slug).first():
            slug = f"{base_slug}-{counter}"
            counter += 1
        return slug

# Event listener to generate slug before insert
@event.listens_for(Project, 'before_insert')
def receive_before_insert(mapper, connection, target):
    if not target.slug:
        target.slug = target._generate_unique_slug(target.title)

# Event listener to generate slug before update (if title changes and slug is not set)
@event.listens_for(Project, 'before_update')
def receive_before_update(mapper, connection, target):
    # Check if the title has changed and slug is not explicitly set
    if target.title != target._sa_instance_state.original.get('title') and not target.slug:
        target.slug = target._generate_unique_slug(target.title)


class AutomationLog(db.Model):
    __tablename__ = 'automation_logs'
    id = db.Column(db.Integer, primary_key=True)
    script_name = db.Column(db.String(100), nullable=False)
    status = db.Column(db.String(50), default='COMPLETE')
    timestamp = db.Column(db.DateTime, default=datetime.utcnow)

class Analytics(db.Model):
    __tablename__ = 'analytics'
    id = db.Column(db.Integer, primary_key=True)
    page_path = db.Column(db.String(255))
    visitor_ip = db.Column(db.String(50))
    timestamp = db.Column(db.DateTime, default=datetime.utcnow)

class User(db.Model, UserMixin):
    __tablename__ = 'login'
    id = db.Column(db.Integer, primary_key=True)
    username = db.Column(db.String(50), unique=True, nullable=False)
    password_hash = db.Column(db.String(255), nullable=False)
