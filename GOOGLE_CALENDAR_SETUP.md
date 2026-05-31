# Google Calendar Booking Setup

This guide connects the custom `/book` page to your Google Calendar so the site can:

- Read your busy events.
- Hide unavailable times with the configured buffer before and after events.
- Create the meeting on your calendar.
- Send the client a calendar invite.
- Add reminders and optionally create a Google Meet link.

## 1. Install Dependencies

Install the new Google Calendar packages:

```bash
pip install -r requirements.txt
```

## 2. Create a Google Cloud Project

1. Open Google Cloud Console.
2. Create a new project or select an existing one.
3. Enable the **Google Calendar API** for that project.

## 3. Create a Service Account

1. Go to **IAM & Admin > Service Accounts**.
2. Create a service account, for example:

```text
jo4dev-booking-calendar
```

3. After creating it, open the service account.
4. Go to **Keys**.
5. Create a new key.
6. Choose **JSON**.
7. Download the JSON file.

## 4. Store the JSON Key

Place the downloaded JSON file inside the app instance folder:

```text
instance/google-service-account.json
```

Do not commit this file to Git.

## 5. Share Your Google Calendar

The service account has an email address inside the JSON file, usually like:

```text
something@project-name.iam.gserviceaccount.com
```

In Google Calendar:

1. Open your calendar settings.
2. Choose the calendar you want bookings added to.
3. Go to **Share with specific people or groups**.
4. Add the service account email.
5. Give it **Make changes to events** permission.

## 6. Get Your Calendar ID

In the same Google Calendar settings page:

1. Scroll to **Integrate calendar**.
2. Copy the **Calendar ID**.

For a personal calendar, this is often your email address. For a secondary calendar, it may look like:

```text
abc123@group.calendar.google.com
```

## 7. Configure Environment Variables

Update `.env`:

```env
GOOGLE_CALENDAR_ID='your-calendar-id-here'
GOOGLE_SERVICE_ACCOUNT_FILE='instance/google-service-account.json'
```

These are fallback values. You can also configure them from the admin panel.

## 8. Configure Booking Settings in Admin

Go to:

```text
/admin/booking
```

Set:

- **Google Calendar ID**: your copied calendar ID.
- **Service Account JSON Path**: `instance/google-service-account.json`.
- **Workday Start / End**: when bookings are allowed.
- **Meeting Minutes**: default is `30`.
- **Buffer Before/After**: default is `30`, blocking 30 minutes before and after each busy event.
- **Slot Step**: how often slots appear, for example every `30` minutes.
- **Book Ahead Days**: how far into the future people can book.
- **Minimum Notice Hours**: prevents last-minute bookings.
- **Reminder Minutes**: calendar reminder timing.
- **Create Google Meet**: enables automatic Meet link creation.

Save the settings.

## 9. Test the Flow

1. Open:

```text
/book
```

2. Click a day on the calendar.
3. Confirm that busy times are hidden.
4. Submit a test booking.
5. Check that:
   - A lead appears in `/admin/leads`.
   - A calendar event appears in your Google Calendar.
   - The client email receives an invite.
   - The event includes reminders.

## Troubleshooting

### Google API Shows Missing

Run:

```bash
pip install -r requirements.txt
```

Then restart the app.

### Slots Show But Calendar Is Not Blocking Busy Times

Check:

- The service account JSON path is correct.
- The calendar is shared with the service account email.
- The service account has **Make changes to events** permission.
- The Calendar ID in `/admin/booking` is correct.

### Booking Saves But No Calendar Event Is Created

The site falls back to a Google Calendar add-event link when credentials are missing or invalid.

Check the response or server logs for the calendar error, then verify the JSON path and calendar sharing.

### Google Meet Link Is Not Created

Make sure **Create Google Meet** is enabled in `/admin/booking`.

Some Google Workspace policies can restrict conference creation. If that happens, disable Google Meet and use a fallback location.

## Security Notes

- Never commit `instance/google-service-account.json`.
- Keep `.env` private.
- Use a dedicated calendar if you do not want the service account to access your main personal calendar.
- The service account only sees calendars you explicitly share with it.
