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
    if not target.slug:
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
    if not target.slug:
        target.slug = target._generate_unique_slug(target.title)


class FAQ(db.Model):
    __tablename__ = 'faqs'
    id = db.Column(db.Integer, primary_key=True)
    question = db.Column(db.String(255), nullable=False)
    slug = db.Column(db.String(255), unique=True, nullable=False, index=True)
    answer = db.Column(db.Text, nullable=False)
    display_order = db.Column(db.Integer, default=0)
    is_published = db.Column(db.Boolean, default=True)
    created_at = db.Column(db.DateTime, default=datetime.utcnow)
    updated_at = db.Column(db.DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        if not self.slug and self.question:
            self.slug = self._generate_unique_slug(self.question)

    def _generate_unique_slug(self, question):
        if not question:
            return ""
        base_slug = make_slug(question)
        slug = base_slug
        counter = 1
        while FAQ.query.filter_by(slug=slug).first():
            slug = f"{base_slug}-{counter}"
            counter += 1
        return slug


@event.listens_for(FAQ, 'before_insert')
def receive_faq_before_insert(mapper, connection, target):
    if not target.slug:
        target.slug = target._generate_unique_slug(target.question)


@event.listens_for(FAQ, 'before_update')
def receive_faq_before_update(mapper, connection, target):
    if not target.slug:
        target.slug = target._generate_unique_slug(target.question)


class BookingSettings(db.Model):
    __tablename__ = 'booking_settings'
    id = db.Column(db.Integer, primary_key=True)
    calendar_id = db.Column(db.String(255), default='primary')
    service_account_file = db.Column(db.String(500))
    workday_start = db.Column(db.String(5), default='09:00')
    workday_end = db.Column(db.String(5), default='17:00')
    meeting_duration_minutes = db.Column(db.Integer, default=30)
    buffer_minutes = db.Column(db.Integer, default=30)
    slot_step_minutes = db.Column(db.Integer, default=30)
    booking_horizon_days = db.Column(db.Integer, default=21)
    min_notice_hours = db.Column(db.Integer, default=4)
    reminder_minutes = db.Column(db.Integer, default=30)
    create_google_meet = db.Column(db.Boolean, default=True)
    meeting_location = db.Column(db.String(255), default='Google Meet')
    updated_at = db.Column(db.DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)


def get_booking_settings():
    settings = BookingSettings.query.first()
    if settings:
        return settings

    settings = BookingSettings()
    db.session.add(settings)
    db.session.commit()
    return settings


class InvoiceSettings(db.Model):
    __tablename__ = 'invoice_settings'
    id = db.Column(db.Integer, primary_key=True)
    biz_name = db.Column(db.String(200), default='JO4 Dev')
    biz_address = db.Column(db.String(400), default='')
    biz_phone = db.Column(db.String(50), default='')
    biz_email = db.Column(db.String(120), default='jroetscyber@gmail.com')
    bank_name = db.Column(db.String(200), default='')
    account_holder = db.Column(db.String(200), default='')
    account_number = db.Column(db.String(100), default='')
    branch_code = db.Column(db.String(50), default='')
    vat_number = db.Column(db.String(50), default='')
    payment_terms = db.Column(db.Text, default='')
    updated_at = db.Column(db.DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)


def get_invoice_settings():
    s = InvoiceSettings.query.first()
    if s:
        return s
    s = InvoiceSettings()
    db.session.add(s)
    db.session.commit()
    return s


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
    created_at = db.Column(db.DateTime, default=datetime.utcnow())

class Service(db.Model):
    __tablename__ = 'services'
    id = db.Column(db.Integer, primary_key=True)
    title = db.Column(db.String(100), nullable=False)
    slug = db.Column(db.String(100), unique=True, nullable=False, index=True)
    eyebrow = db.Column(db.String(50))
    lead_text = db.Column(db.Text)
    description = db.Column(db.Text)
    features = db.Column(db.Text) # Newline separated
    price_range = db.Column(db.String(100))
    price_label = db.Column(db.String(100))
    price_note = db.Column(db.String(255))
    icon_svg = db.Column(db.Text) # The SVG path or full SVG code
    panel_title = db.Column(db.String(100))
    panel_type = db.Column(db.String(50)) # 'use-case', 'audit', 'stats', 'auto'
    panel_content = db.Column(db.Text) # JSON or structured text
    is_published = db.Column(db.Boolean, default=True)
    display_order = db.Column(db.Integer, default=0)
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
        while Service.query.filter_by(slug=slug).first():
            slug = f"{base_slug}-{counter}"
            counter += 1
        return slug

@event.listens_for(Service, 'before_insert')
def receive_service_before_insert(mapper, connection, target):
    if not target.slug:
        target.slug = target._generate_unique_slug(target.title)

@event.listens_for(Service, 'before_update')
def receive_service_before_update(mapper, connection, target):
    if not target.slug:
        target.slug = target._generate_unique_slug(target.title)

class FAQSubmission(db.Model):
    __tablename__ = 'faq_submissions'
    id = db.Column(db.Integer, primary_key=True)
    name = db.Column(db.String(100), nullable=False)
    email = db.Column(db.String(100), nullable=False)
    phone = db.Column(db.String(50))
    question = db.Column(db.Text, nullable=False)
    created_at = db.Column(db.DateTime, default=datetime.utcnow)
    is_answered = db.Column(db.Boolean, default=False)
