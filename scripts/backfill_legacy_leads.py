from app import app
from legacy_leads import backfill_legacy_leads


if __name__ == "__main__":
    backfill_legacy_leads(app)
