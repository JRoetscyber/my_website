from datetime import datetime

from database import db, Lead
from lead_score import calculate_lead_score
from admin import normalize_lead_status


def backfill_legacy_leads(app, verbose=True):
    updated = 0

    with app.app_context():
        leads = Lead.query.order_by(Lead.id.asc()).all()

        for lead in leads:
            needs_backfill = any([
                lead.status is None,
                lead.score is None,
                lead.project_type is None,
                lead.explicit_score is None,
                lead.implicit_score is None,
                lead.urgency_score is None,
                lead.last_activity_date is None,
            ])

            if not needs_backfill:
                continue

            last_activity = lead.last_activity_date or lead.created_at or datetime.utcnow()
            scoring_input = {
                "client_name": lead.client_name or lead.client_company or "Legacy Lead",
                "client_company": lead.client_company,
                "contact_role": lead.contact_role or "Agent",
                "budget": lead.budget or 0,
                "project_type": lead.project_type or lead.target_project or "Static",
                "visited_pages": ["/admin"],
                "whatsapp_engagement": lead.whatsapp_engagement or "Ignored",
                "phone_number": lead.phone_number,
                "last_activity_date": last_activity.strftime("%Y-%m-%d"),
            }

            result = calculate_lead_score(scoring_input)

            lead.project_type = lead.project_type or lead.target_project or "Static"
            lead.target_project = lead.target_project or lead.project_type
            lead.status = normalize_lead_status(lead.status or "New")
            lead.score = int(round(result["score"]))
            lead.explicit_score = result["breakdown"]["explicit"]
            lead.implicit_score = result["breakdown"]["implicit"]
            lead.urgency_score = result["breakdown"]["urgency"]
            lead.last_activity_date = last_activity
            updated += 1

        db.session.commit()

    if verbose:
        print(f"Backfilled {updated} legacy leads.")

    return updated
