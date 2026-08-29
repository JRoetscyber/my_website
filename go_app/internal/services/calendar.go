package services

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/JRoetscyber/my_website/go_app/internal/models"
)

var SAST = time.FixedZone("SAST", 2*60*60)

type TimeSlot struct {
	Value  string `json:"value"`
	Label  string `json:"label"`
	EndsAt string `json:"ends_at"`
}

type AvailabilityResult struct {
	Status                   string                 `json:"status"`
	Date                     string                 `json:"date"`
	Slots                    []TimeSlot             `json:"slots"`
	Busy                     []map[string]string    `json:"busy"`
	CalendarConnected        bool                   `json:"calendar_connected"`
	CalendarError            *string                `json:"calendar_error"`
	GooglePackagesInstalled  bool                   `json:"google_packages_installed"`
	Settings                 map[string]interface{} `json:"settings"`
}

func parseTimePart(value string, defHour, defMin int) (int, int) {
	parts := strings.Split(value, ":")
	if len(parts) >= 2 {
		h, err1 := strconv.Atoi(parts[0])
		m, err2 := strconv.Atoi(parts[1])
		if err1 == nil && err2 == nil {
			return h, m
		}
	}
	return defHour, defMin
}

func BuildAvailableSlots(settings *models.BookingSettings, dateValue string) AvailabilityResult {
	parsedDate, err := time.ParseInLocation("2006-01-02", dateValue, SAST)
	if err != nil {
		parsedDate = time.Now().In(SAST)
		dateValue = parsedDate.Format("2006-01-02")
	}

	startH, startM := parseTimePart(settings.WorkdayStart, 9, 0)
	endH, endM := parseTimePart(settings.WorkdayEnd, 17, 0)

	workStart := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), startH, startM, 0, 0, SAST)
	workEnd := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), endH, endM, 0, 0, SAST)

	durationM := settings.MeetingDurationMinutes
	if durationM <= 0 {
		durationM = 30
	}
	stepM := settings.SlotStepMinutes
	if stepM <= 0 {
		stepM = 30
	}
	noticeH := settings.MinNoticeHours
	if noticeH <= 0 {
		noticeH = 4
	}

	minStart := time.Now().In(SAST).Add(time.Duration(noticeH) * time.Hour)
	slots := make([]TimeSlot, 0)

	cursor := workStart
	for cursor.Add(time.Duration(durationM)*time.Minute).Before(workEnd) || cursor.Add(time.Duration(durationM)*time.Minute).Equal(workEnd) {
		slotEnd := cursor.Add(time.Duration(durationM) * time.Minute)
		// Only Monday-Friday (Weekday 1..5)
		if cursor.After(minStart) && cursor.Weekday() >= time.Monday && cursor.Weekday() <= time.Friday {
			slots = append(slots, TimeSlot{
				Value:  cursor.Format(time.RFC3339),
				Label:  cursor.Format("15:04"),
				EndsAt: slotEnd.Format(time.RFC3339),
			})
		}
		cursor = cursor.Add(time.Duration(stepM) * time.Minute)
	}

	return AvailabilityResult{
		Status:                  "success",
		Date:                    dateValue,
		Slots:                   slots,
		Busy:                    []map[string]string{},
		CalendarConnected:       true,
		GooglePackagesInstalled: true,
		Settings: map[string]interface{}{
			"duration":      durationM,
			"buffer":        settings.BufferMinutes,
			"horizon_days":  settings.BookingHorizonDays,
			"workday_start": settings.WorkdayStart,
			"workday_end":   settings.WorkdayEnd,
		},
	}
}

func GoogleCalendarURL(startTime time.Time, name, company, email, phone, projectType, notes string) string {
	endTime := startTime.Add(30 * time.Minute)
	startUTC := startTime.UTC().Format("20060102T150405Z")
	endUTC := endTime.UTC().Format("20060102T150405Z")

	companyText := company
	if companyText == "" {
		companyText = "Not provided"
	}
	phoneText := phone
	if phoneText == "" {
		phoneText = "Not provided"
	}
	notesText := notes
	if notesText == "" {
		notesText = "No extra notes provided."
	}

	details := fmt.Sprintf("Discovery call booked through JO4 Dev.\nName: %s\nCompany: %s\nEmail: %s\nPhone: %s\nProject type: %s\n\n%s",
		name, companyText, email, phoneText, projectType, notesText)

	subjectText := name
	if company != "" {
		subjectText = company
	}

	params := url.Values{}
	params.Set("action", "TEMPLATE")
	params.Set("text", fmt.Sprintf("JO4 Dev Discovery Call - %s", subjectText))
	params.Set("dates", fmt.Sprintf("%s/%s", startUTC, endUTC))
	params.Set("details", details)
	params.Set("location", "Google Meet / Phone call")
	params.Set("trp", "false")

	return fmt.Sprintf("https://calendar.google.com/calendar/render?%s", params.Encode())
}
