import os
from mailer import send_booking_confirmation

# 1. Force Python to load your .env file
try:
    from dotenv import load_dotenv
    load_dotenv()
except ImportError:
    print("Warning: python-dotenv not installed. If env vars fail, run: pip install python-dotenv")

print("Checking environment variables...")
print(f"SENDER: {os.getenv('SENDER_EMAIL')}")
print(f"PASSWORD: {'LOADED' if os.getenv('EMAIL_APP_PASSWORD') else 'MISSING'}")

print("\nAttempting to send test email...")
# CHANGE THIS to your personal email address so you can see if it arrives
test_client_email = "YOUR_PERSONAL_EMAIL@gmail.com" 

success = send_booking_confirmation(test_client_email, "API Tester", "Right Now")

if success:
    print("\n✅ Email fired successfully! Check your inbox.")
else:
    print("\n❌ Email failed. Look at the error message above.")