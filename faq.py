from flask import Blueprint, abort, render_template, url_for
from database import FAQ, db
import json


faq_bp = Blueprint('faq', __name__)


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
        "url": url_for('faq.faq_detail', slug=faq.slug, _external=True)
    }

    return render_template(
        'faq_detail.html',
        faq=faq,
        json_ld_schema=json.dumps(schema_data, indent=2)
    )
