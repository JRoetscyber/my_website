package services

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"time"

	"github.com/JRoetscyber/my_website/go_app/internal/config"
)

func GenerateICSAttachment(startTime time.Time) string {
	endTime := startTime.Add(30 * time.Minute)
	dtstamp := time.Now().UTC().Format("20060102T150405Z")
	dtstart := startTime.UTC().Format("20060102T150405Z")
	dtend := endTime.UTC().Format("20060102T150405Z")

	ics := fmt.Sprintf("BEGIN:VCALENDAR\r\n"+
		"VERSION:2.0\r\n"+
		"PRODID:-//JO4 Dev//Booking System//EN\r\n"+
		"CALSCALE:GREGORIAN\r\n"+
		"METHOD:REQUEST\r\n"+
		"BEGIN:VEVENT\r\n"+
		"DTSTAMP:%s\r\n"+
		"DTSTART:%s\r\n"+
		"DTEND:%s\r\n"+
		"SUMMARY:Discovery Call with JO4 Dev\r\n"+
		"DESCRIPTION:Your 30-minute discovery call is confirmed. Looking forward to our chat!\r\n"+
		"STATUS:CONFIRMED\r\n"+
		"SEQUENCE:0\r\n"+
		"END:VEVENT\r\n"+
		"END:VCALENDAR\r\n", dtstamp, dtstart, dtend)

	return ics
}

func SendBookingInvite(cfg *config.Config, clientEmail, clientName string, scheduledAt time.Time) error {
	if cfg.MailUsername == "" || cfg.MailPassword == "" {
		log.Printf("[Mailer] SMTP credentials not set. Simulating booking email to %s for %s", clientEmail, scheduledAt.Format(time.RFC1123))
		return nil
	}

	subject := "Your Discovery Call is Confirmed! | JO4 Dev"
	icsData := GenerateICSAttachment(scheduledAt)

	boundary := "===JO4_MAIL_BOUNDARY==="

	body := fmt.Sprintf("From: JO4 Dev <%s>\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: multipart/mixed; boundary=%s\r\n\r\n"+
		"--%s\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n\r\n"+
		`<!DOCTYPE html>
<html>
<head>
<style>
body { font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif; background-color: #f4f7f6; color: #333333; margin: 0; padding: 0; }
.container { max-width: 600px; margin: 40px auto; background: #ffffff; border-radius: 8px; overflow: hidden; box-shadow: 0 4px 15px rgba(0,0,0,0.05); }
.header { background-color: #1a1a1a; color: #ffffff; padding: 30px 40px; text-align: center; }
.header h1 { margin: 0; font-size: 24px; font-weight: 600; letter-spacing: 1px; }
.content { padding: 40px; line-height: 1.6; }
.content p { margin-bottom: 20px; font-size: 16px; color: #444444; }
.highlight { background-color: #f8f9fa; padding: 20px; border-left: 4px solid #0066cc; margin-bottom: 25px; border-radius: 4px; }
.highlight p { margin: 0; font-size: 15px; color: #333333; }
.footer { background-color: #f9fbfb; padding: 20px 40px; text-align: center; font-size: 14px; color: #888888; border-top: 1px solid #eeeeee; }
</style>
</head>
<body>
<div class="container">
<div class="header">
<h1>JO4 DEV</h1>
</div>
<div class="content">
<p>Hi <strong>%s</strong>,</p>
<p>Your 30-minute discovery call is successfully booked!</p>
<div class="highlight">
<p>📅 <strong>Action Required:</strong> Please open the attached calendar invitation (<code>invite.ics</code>) to automatically add this meeting to your schedule.</p>
</div>
<p>If you need to reschedule or share any extra notes with me before our call, simply reply to <a href="mailto:info@jo4.co.za" style="color: #d32f2f; text-decoration: none; font-weight: bold;">info@jo4.co.za</a>.</p>
<p>Looking forward to chatting about your project!</p>
<p style="margin-top: 40px;">Speak soon,<br><strong>Jonathan</strong><br>JO4 Dev</p>
</div>
<div class="footer">
&copy; %d JO4 Dev. All rights reserved.
</div>
</div>
</body>
</html>`+
		"\r\n\r\n--%s\r\n"+
		"Content-Type: text/calendar; name=\"invite.ics\"; method=REQUEST\r\n"+
		"Content-Disposition: attachment; filename=\"invite.ics\"\r\n"+
		"Content-Transfer-Encoding: 8bit\r\n\r\n"+
		"%s\r\n"+
		"--%s--",
		cfg.MailSender, clientEmail, subject, boundary, boundary,
		clientName, time.Now().Year(), boundary, icsData, boundary)

	auth := smtp.PlainAuth("", cfg.MailUsername, cfg.MailPassword, cfg.MailServer)
	addr := fmt.Sprintf("%s:%s", cfg.MailServer, cfg.MailPort)

	// If port 465 (SSL), use TLS connection, else STARTTLS / standard
	if strings.Contains(cfg.MailPort, "465") {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         cfg.MailServer,
		}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return err
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, cfg.MailServer)
		if err != nil {
			return err
		}
		defer client.Close()

		if err = client.Auth(auth); err != nil {
			return err
		}
		if err = client.Mail(cfg.MailSender); err != nil {
			return err
		}
		if err = client.Rcpt(clientEmail); err != nil {
			return err
		}
		w, err := client.Data()
		if err != nil {
			return err
		}
		_, err = w.Write([]byte(body))
		if err != nil {
			return err
		}
		return w.Close()
	}

	return smtp.SendMail(addr, auth, cfg.MailSender, []string{clientEmail}, []byte(body))
}
