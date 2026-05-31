from flask import Blueprint, abort, render_template, url_for, request, redirect, current_app
from database import FAQ, FAQSubmission, db
from flask_mail import Message
import json


faq_bp = Blueprint('faq', __name__)
...
@faq_bp.route('/faq/submit', methods=['POST'])
def submit_question():
    try:
        name = request.form.get('name')
        email = request.form.get('email')
        phone = request.form.get('phone')
        question = request.form.get('question')

        if not name or not email or not question:
            return redirect(url_for('faq.faq_list'))

        new_sub = FAQSubmission(
            name=name,
            email=email,
            phone=phone,
            question=question
        )
        db.session.add(new_sub)
        db.session.commit()

        # Notify Admin
        try:
            mail = current_app.extensions.get('mail')
            msg = Message(
                subject=f"New FAQ Question from {name}",
                recipients=[current_app.config['MAIL_DEFAULT_SENDER']],
                body=f"Name: {name}\nEmail: {email}\nPhone: {phone}\n\nQuestion:\n{question}"
            )
            mail.send(msg)
        except Exception as e:
            print(f"Error sending notification email: {e}")

    except Exception as e:
        db.session.rollback()
        current_app.logger.exception("Error submitting FAQ question: %s", e)
    
    return redirect(url_for('faq.faq_list'))


def published_faqs_query():
    return FAQ.query.filter_by(is_published=True).order_by(FAQ.display_order.asc(), FAQ.created_at.desc())


@faq_bp.route('/faq')
def faq_list():
    try:
        faqs = published_faqs_query().all()
    except Exception as e:
        db.session.rollback()
        faqs = []
        print(f"Error fetching FAQs: {e}")

    schema_data = {
        "@context": "https://schema.org",
        "@type": "FAQPage",
        "mainEntity": [
            {
                "@type": "Question",
                "name": faq.question,
                "acceptedAnswer": {
                    "@type": "Answer",
                    "text": faq.answer
                }
            }
            for faq in faqs
        ]
    }

    return render_template(
        'faq.html',
        faqs=faqs,
        json_ld_schema=json.dumps(schema_data, indent=2)
    )


@faq_bp.route('/faq/<slug>')
def faq_detail(slug):
    faq = FAQ.query.filter_by(slug=slug, is_published=True).first_or_404()
    schema_data = {
        "@context": "https://schema.org",
        "@type": "Question",
        "name": faq.question,
        "acceptedAnswer": {
            "@type": "Answer",
            "text": faq.answer
        },
        "url": url_for('faq.faq_detail', slug=faq.slug, _external=True, _scheme='https')
    }

    return render_template(
        'faq_detail.html',
        faq=faq,
        json_ld_schema=json.dumps(schema_data, indent=2)
    )
