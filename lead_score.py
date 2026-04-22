import math
from datetime import datetime, timedelta

def calculate_lead_score(lead_data):
    """
    Lead Scoring Engine v1.0
    
    Weights Breakdown:
    1. Explicit Score (40%): Structural fit (Company size, Role, Project Type)
    2. Implicit Score (40%): Engagement signals (Web analytics, WhatsApp behavior, Data completeness)
    3. Urgency Score  (20%): Time-decay based on the 'last_activity' timestamp.
    
    Classification:
    - Hot: 80-100
    - Warm: 50-79
    - Cold: 0-49
    """
    
    # 1. EXPLICIT DATA SCORING (Max 100)
    # -------------------------------------------------------------------------
    explicit_score = 0
    
    # Budget Scoring (33.3% of Explicit)
    # Scale: R1,000 (10 pts) to R500,000 (100 pts)
    raw_budget = lead_data.get('budget', 0)
    try:
        budget = float(raw_budget)
    except (ValueError, TypeError):
        budget = 0

    if budget >= 500000:
        budget_points = 100
    elif budget <= 1000:
        budget_points = 10
    else:
        # Linear interpolation between 1k and 500k
        budget_points = 10 + ((budget - 1000) / (500000 - 1000) * 90)
    
    explicit_score += budget_points * 0.333
    
    # Decision Maker Status (33.3% of Explicit)
    role_map = {'owner': 100, 'principal': 100, 'director': 80, 'manager': 50, 'agent': 20}
    explicit_score += role_map.get(lead_data.get('contact_role', '').lower(), 10) * 0.333
    
    # Project Alignment (33.3% of Explicit)
    # Portals/ERPs are high-value target niches
    project_map = {
    'erp': 100,              # Full inventory, stock taking, and sales prediction systems
    'portal': 100,           # Secure client login areas (e.g., buyer/valuation portals)
    'iot_integration': 90,   # Dashboards for embedded systems, long-range trackers, or sensor data
    'data_pipeline': 80,     # Scripts that clean or merge large databases into usable formats
    'crm': 70,               # Custom lead tracking and pipeline management
    'server_config': 70,     # Deploying secure, remote-accessible local web servers via tunneling
    'automation': 60,        # Scheduled messaging, automated outreach, or report generation
    'pos_system': 60,        # Point of Sale integrations for mobile or retail businesses
    'static': 20             # Basic HTML/CSS informational pages
    }
    explicit_score += project_map.get(lead_data.get('project_type', '').lower(), 10) * 0.334

    # 2. IMPLICIT DATA SCORING (Max 100)
    # -------------------------------------------------------------------------
    implicit_score = 0
    
    # Website Analytics (40% of Implicit)
    high_intent_pages = ['/projects', '/book', '/pricing', 'case-study']
    viewed_pages = lead_data.get('visited_pages', [])
    intent_match = any(page in viewed_pages for page in high_intent_pages)
    implicit_score += (100 if intent_match else 20) * 0.40
    
    # Outreach Responsiveness (40% of Implicit)
    # WhatsApp metrics: 'replied' is the gold standard
    whatsapp_status = lead_data.get('whatsapp_engagement', 'ignored').lower()
    wa_map = {'replied': 100, 'read': 50, 'ignored': 0}
    implicit_score += wa_map.get(whatsapp_status, 0) * 0.40
    
    # Data Completeness (20% of Implicit)
    # Bonus for providing phone + company name
    completeness = 0
    if lead_data.get('client_company'): completeness += 50
    if lead_data.get('phone_number'): completeness += 50
    implicit_score += completeness * 0.20

    # 3. TIME DECAY / URGENCY (Max 100)
    # -------------------------------------------------------------------------
    # Formula: Score = Max_Score * e^(-k * t)
    # where t = days since activity - grace_period
    now = datetime.now()
    last_act = lead_data.get('last_activity_date', now)
    if isinstance(last_act, str):
        last_act = datetime.strptime(last_act, '%Y-%m-%d')
        
    days_since = (now - last_act).days
    grace_period = 14 # No decay for first 14 days
    
    if days_since <= grace_period:
        urgency_score = 100
    else:
        # Decay constant k = 0.1 gives a ~50% drop every 7 days after the grace period
        k = 0.1
        t = days_since - grace_period
        urgency_score = 100 * math.exp(-k * t)

    # FINAL AGGREGATION
    # -------------------------------------------------------------------------
    final_score = (
        (explicit_score * 0.40) + 
        (implicit_score * 0.40) + 
        (urgency_score  * 0.20)
    )
    
    final_score = round(final_score, 1)
    
    # Classification
    if final_score >= 80:
        classification = 'Hot'
    elif final_score >= 50:
        classification = 'Warm'
    else:
        classification = 'Cold'
        
    return {
        "score": final_score,
        "classification": classification,
        "breakdown": {
            "explicit": round(explicit_score, 1),
            "implicit": round(implicit_score, 1),
            "urgency": round(urgency_score, 1)
        }
    }

# --- TEST CASES ---
if __name__ == "__main__":
    now_str = datetime.now().strftime('%Y-%m-%d')
    old_str = (datetime.now() - timedelta(days=25)).strftime('%Y-%m-%d')

    # Case 1: High-intent principal (Hot)
    lead_1 = {
        "client_company": "Sandton Realty Group",
        "contact_role": "Principal",
        "budget_tier": "large",
        "project_type": "portal",
        "visited_pages": ["/projects", "/book"],
        "whatsapp_engagement": "replied",
        "phone_number": "+27 82 123 4567",
        "last_activity_date": now_str
    }
    
    # Case 2: Mid-level manager, low engagement, old activity (Warm/Cold)
    lead_2 = {
        "client_company": "Cape Town Rentals",
        "contact_role": "Manager",
        "budget_tier": "medium",
        "project_type": "static",
        "visited_pages": ["/"],
        "whatsapp_engagement": "read",
        "phone_number": None,
        "last_activity_date": old_str # 25 days ago
    }
    
    for i, lead in enumerate([lead_1, lead_2], 1):
        result = calculate_lead_score(lead)
        print(f"--- Lead Score Report #{i} ---")
        print(f"Lead: {lead['client_company']}")
        print(f"Final Score: {result['score']}/100")
        print(f"Classification: {result['classification']}")
        print(f"Breakdown: {result['breakdown']}\n")
