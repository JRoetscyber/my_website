from flask_sqlalchemy import SQLAlchemy
from datetime import datetime
from flask_login import UserMixin
from sqlalchemy import event
import re
import unicodedata

db = SQLAlchemy()

def make_slug(value):
    value = str(value or '')
    value = unicodedata.normalize('NFKD', value).encode('ascii', 'ignore').decode('ascii')
    value = value.lower()
    value = re.sub(r'[^a-z0-9]+', '-', value).strip('-')
    return value or 'project'

class Lead(db.Model):
    __tablename__ = 'leads'
    id = db.Column(db.Integer, primary_key=True)
    client_name = db.Column(db.String(100), nullable=False)
    client_company = db.Column(db.String(100))
    project_type = db.Column(db.String(100))
    budget = db.Column(db.Float)
    contact_role = db.Column(db.String(100))
    phone_number = db.Column(db.String(50))
    whatsapp_engagement = db.Column(db.String(50))
    target_project = db.Column(db.String(100))
    score = db.Column(db.Integer)
    explicit_score = db.Column(db.Float)
    implicit_score = db.Column(db.Float)
    urgency_score = db.Column(db.Float)
    status = db.Column(db.String(50), default='New') # 'New', 'Contacted', 'Proposal Sent', 'Closed'
    loss_reason = db.Column(db.String(255))
    created_at = db.Column(db.DateTime, default=datetime.utcnow)
    last_activity_date = db.Column(db.DateTime, default=datetime.utcnow)

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
    performance = db.Column(db.Integer, nullable=True)
    seo = db.Column(db.Integer, nullable=True)
    deployed_at = db.Column(db.DateTime, default=datetime.utcnow)

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        if not self.slug:
            self.slug = self._generate_unique_slug(self.title)

    def _generate_unique_slug(self, title):
        if not title:
            return ""
        base_slug = make_slug(title)
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


class BlogPost(db.Model):
    __tablename__ = 'blog_posts'
    id = db.Column(db.Integer, primary_key=True)
    title = db.Column(db.String(200), nullable=False)
    slug = db.Column(db.String(255), unique=True, nullable=False, index=True)
    summary = db.Column(db.Text)
    content = db.Column(db.Text, nullable=False)
    media_path = db.Column(db.String(255))
    views = db.Column(db.Integer, default=0)
    created_at = db.Column(db.DateTime, default=datetime.utcnow)
    updated_at = db.Column(db.DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        if not self.slug and self.title:
            self.slug = self._generate_unique_slug(self.title)

    def _generate_unique_slug(self, title):
        if not title:
            return ""
        base_slug = make_slug(title)
        slug = base_slug
        counter = 1
        while BlogPost.query.filter_by(slug=slug).first():
            slug = f"{base_slug}-{counter}"
            counter += 1
        return slug

# Event listeners for BlogPost slugs
@event.listens_for(BlogPost, 'before_insert')
def receive_blog_before_insert(mapper, connection, target):
    if not target.slug:
        target.slug = target._generate_unique_slug(target.title)

@event.listens_for(BlogPost, 'before_update')
def receive_blog_before_update(mapper, connection, target):
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


class Transaction(db.Model):
    __tablename__ = 'transactions'
    id = db.Column(db.Integer, primary_key=True)
    type = db.Column(db.String(20), nullable=False) # 'Income' or 'Expense'
    category = db.Column(db.String(100))
    amount = db.Column(db.Float, nullable=False)
    description = db.Column(db.Text)
    date = db.Column(db.Date, default=datetime.utcnow().date())
    created_at = db.Column(db.DateTime, default=datetime.utcnow)
