from flask_mail import Message
from flask import current_app
from datetime import timedelta, timezone

def generate_ics_attachment(start_time):
    """Generates the raw text for a standard calendar invitation file."""
    end_time = start_time + timedelta(minutes=30)
    
    # .ics files require strict UTC time formatting
    dtstamp = start_time.now(timezone.utc).strftime('%Y%m%dT%H%M%SZ')
    dtstart = start_time.astimezone(timezone.utc).strftime('%Y%m%dT%H%M%SZ')
    dtend = end_time.astimezone(timezone.utc).strftime('%Y%m%dT%H%M%SZ')

    ics_content = f"""BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//JO4 Dev//Booking System//EN
CALSCALE:GREGORIAN
METHOD:REQUEST
BEGIN:VEVENT
DTSTAMP:{dtstamp}
DTSTART:{dtstart}
DTEND:{dtend}
SUMMARY:Discovery Call with JO4 Dev
DESCRIPTION:Your 30-minute discovery call is confirmed. Looking forward to our chat!
STATUS:CONFIRMED
SEQUENCE:0
END:VEVENT
END:VCALENDAR"""

    # .ics files require carriage returns (\r\n) for new lines
    return ics_content.replace('\n', '\r\n')

def send_booking_invite(client_email, client_name, scheduled_at):
    """Sends the email with the .ics calendar attachment and professional HTML layout."""
    
    msg = Message(
        subject="Your Discovery Call is Confirmed! | JO4 Dev",
        recipients=[client_email]
    )
    
    # 1. Plain Text Fallback (Important for spam filters)
    msg.body = f"""Hi {client_name},

Your 30-minute discovery call is successfully booked!

Please open the attached calendar invitation (.ics) to automatically add this meeting to your schedule. 

If you need to share any extra notes before the call, just reply directly to this email.

Speak soon,
Jonathan
JO4 Dev"""

    # 2. Professional HTML Version
    msg.html = f"""
    <!DOCTYPE html>
    <html>
    <head>
        <style>
            body {{ font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif; background-color: #f4f7f6; color: #333333; margin: 0; padding: 0; }}
            .container {{ max-width: 600px; margin: 40px auto; background: #ffffff; border-radius: 8px; overflow: hidden; box-shadow: 0 4px 15px rgba(0,0,0,0.05); }}
            .header {{ background-color: #1a1a1a; color: #ffffff; padding: 30px 40px; text-align: center; }}
            .header h1 {{ margin: 0; font-size: 24px; font-weight: 600; letter-spacing: 1px; }}
            .content {{ padding: 40px; line-height: 1.6; }}
            .content p {{ margin-bottom: 20px; font-size: 16px; color: #444444; }}
            .highlight {{ background-color: #f8f9fa; padding: 20px; border-left: 4px solid #0066cc; margin-bottom: 25px; border-radius: 4px; }}
            .highlight p {{ margin: 0; font-size: 15px; color: #333333; }}
            .footer {{ background-color: #f9fbfb; padding: 20px 40px; text-align: center; font-size: 14px; color: #888888; border-top: 1px solid #eeeeee; }}
        </style>
    </head>
    <body>
        <div class="container">
            <div class="header">
                <h1>JO4 DEV</h1>
            </div>
            <div class="content">
                <p>Hi <strong>{client_name}</strong>,</p>
                <p>Your 30-minute discovery call is successfully booked!</p>
                
                <div class="highlight">
                    <p>📅 <strong>Action Required:</strong> Please open the attached calendar invitation (<code>invite.ics</code>) to automatically add this meeting and the connection details to your schedule.</p>
                </div>
                
                <p>If you need to reschedule or share any extra notes with me before our call, simply reply to <a href="mailto:info@jo4.co.za" style="color: #d32f2f; text-decoration: none; font-weight: bold;">info@jo4.co.za</a>.</p>
                <p>Looking forward to chatting about your project!</p>
                
                <p style="margin-top: 40px;">Speak soon,<br><strong>Jonathan</strong><br>JO4 Dev</p>
            </div>
            <div class="footer">
                &copy; 2026 JO4 Dev. All rights reserved.
            </div>
        </div>
    </body>
    </html>
    """

    # 3. Generate the calendar file and attach it
    ics_data = generate_ics_attachment(scheduled_at)
    msg.attach("invite.ics", "text/calendar", ics_data)

    # 4. Send using the mail instance from the current running Flask app
    try:
        mail = current_app.extensions.get('mail')
        mail.send(msg)
        print(f"Success: Professional HTML invite sent to {client_email} via Brevo")
        return True
    except Exception as e:
        print(f"Error sending email via Brevo: {e}")
        return False