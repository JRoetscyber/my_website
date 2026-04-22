from flask_sqlalchemy import SQLAlchemy
from datetime import datetime

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
    category = db.Column(db.String(50))
    tech_stack = db.Column(db.Text) # Stored as comma-separated string
    description = db.Column(db.Text)
    code_snippet = db.Column(db.Text)
    media_path = db.Column(db.String(255))
    views = db.Column(db.Integer, default=0)
    deployed_at = db.Column(db.DateTime, default=datetime.utcnow)

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

class User(db.Model):
    __tablename__ = 'login'
    id = db.Column(db.Integer, primary_key=True)
    username = db.Column(db.String(50), unique=True, nullable=False)
    password_hash = db.Column(db.String(255), nullable=False)
