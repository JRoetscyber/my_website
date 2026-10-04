package models

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

type Lead struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	ClientName         string    `gorm:"size:100;not null" json:"client_name"`
	ClientCompany      string    `gorm:"size:100" json:"client_company"`
	ProjectType        string    `gorm:"size:100" json:"project_type"`
	Budget             float64   `json:"budget"`
	ContactRole        string    `gorm:"size:100" json:"contact_role"`
	PhoneNumber        string    `gorm:"size:50" json:"phone_number"`
	WhatsappEngagement string    `gorm:"size:50" json:"whatsapp_engagement"`
	TargetProject      string    `gorm:"size:100" json:"target_project"`
	Score              int       `json:"score"`
	ExplicitScore      float64   `json:"explicit_score"`
	ImplicitScore      float64   `json:"implicit_score"`
	UrgencyScore       float64   `json:"urgency_score"`
	Status             string    `gorm:"size:50" json:"status"`
	LossReason         string    `gorm:"size:255" json:"loss_reason"`
	CreatedAt          time.Time `gorm:"autoCreateTime" json:"created_at"`
	LastActivityDate   time.Time `json:"last_activity_date"`
}

func (Lead) TableName() string {
	return "leads"
}

func (l *Lead) BeforeCreate(tx *gorm.DB) error {
	if l.Status == "" {
		l.Status = "New"
	}
	return nil
}

type Project struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"size:100;not null" json:"title"`
	Slug        string    `gorm:"size:200;uniqueIndex;not null" json:"slug"`
	Category    string    `gorm:"size:50" json:"category"`
	TechStack   string    `gorm:"type:text" json:"tech_stack"`
	Description string    `gorm:"type:text" json:"description"`
	CodeSnippet string    `gorm:"type:text" json:"code_snippet"`
	YoutubeURL  string    `gorm:"size:500" json:"youtube_url"`
	ProjectURL  string    `gorm:"size:500" json:"project_url"`
	MediaPath   string    `gorm:"size:255" json:"media_path"`
	Views       int       `gorm:"default:0" json:"views"`
	Performance *int      `json:"performance"`
	SEO         *int      `json:"seo"`
	DeployedAt  time.Time `gorm:"autoCreateTime" json:"deployed_at"`
}

func (Project) TableName() string {
	return "projects"
}

type BlogPost struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"size:200;not null" json:"title"`
	Slug        string    `gorm:"size:255;uniqueIndex;not null" json:"slug"`
	Summary     string    `gorm:"type:text" json:"summary"`
	Content     string    `gorm:"type:text;not null" json:"content"`
	MediaPath   string    `gorm:"size:255" json:"media_path"`
	Views       int       `gorm:"default:0" json:"views"`
	Status      string    `gorm:"size:50" json:"status"`
	PublishedAt time.Time `json:"published_at"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (BlogPost) TableName() string {
	return "blog_posts"
}

func (b *BlogPost) BeforeCreate(tx *gorm.DB) error {
	if b.Status == "" {
		b.Status = "published"
	}
	return nil
}

type FAQ struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Question     string    `gorm:"size:255;not null" json:"question"`
	Slug         string    `gorm:"size:255;uniqueIndex;not null" json:"slug"`
	Answer       string    `gorm:"type:text;not null" json:"answer"`
	DisplayOrder int       `gorm:"default:0" json:"display_order"`
	IsPublished  bool      `gorm:"default:true" json:"is_published"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (FAQ) TableName() string {
	return "faqs"
}

type Service struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	Title            string    `gorm:"size:100;not null" json:"title"`
	Slug             string    `gorm:"size:100;uniqueIndex;not null" json:"slug"`
	Eyebrow          string    `gorm:"size:50" json:"eyebrow"`
	LeadText         string    `gorm:"type:text" json:"lead_text"`
	Description      string    `gorm:"type:text" json:"description"`
	Features         string    `gorm:"type:text" json:"features"`
	PriceRange       string    `gorm:"size:100" json:"price_range"`
	PriceLabel       string    `gorm:"size:100" json:"price_label"`
	PriceNote        string    `gorm:"size:255" json:"price_note"`
	IconSVG          string    `gorm:"type:text" json:"icon_svg"`
	PanelTitle       string    `gorm:"size:100" json:"panel_title"`
	PanelType        string    `gorm:"size:50" json:"panel_type"`
	PanelContent     string    `gorm:"type:text" json:"panel_content"`
	IsPublished      bool      `gorm:"default:true" json:"is_published"`
	HasDedicatedPage bool      `gorm:"default:false" json:"has_dedicated_page"`
	DisplayOrder     int       `gorm:"default:0" json:"display_order"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Service) TableName() string {
	return "services"
}

func (s Service) FeatureList() []string {
	lines := strings.Split(strings.ReplaceAll(s.Features, "\r\n", "\n"), "\n")
	var res []string
	for _, l := range lines {
		if tr := strings.TrimSpace(l); tr != "" {
			res = append(res, tr)
		}
	}
	return res
}

type BookingSettings struct {
	ID                     uint      `gorm:"primaryKey" json:"id"`
	CalendarID             string    `gorm:"size:255;default:'primary'" json:"calendar_id"`
	ServiceAccountFile     string    `gorm:"size:500" json:"service_account_file"`
	WorkdayStart           string    `gorm:"size:5;default:'09:00'" json:"workday_start"`
	WorkdayEnd             string    `gorm:"size:5;default:'17:00'" json:"workday_end"`
	MeetingDurationMinutes int       `gorm:"default:30" json:"meeting_duration_minutes"`
	BufferMinutes          int       `gorm:"default:30" json:"buffer_minutes"`
	SlotStepMinutes        int       `gorm:"default:30" json:"slot_step_minutes"`
	BookingHorizonDays     int       `gorm:"default:21" json:"booking_horizon_days"`
	MinNoticeHours         int       `gorm:"default:4" json:"min_notice_hours"`
	ReminderMinutes        int       `gorm:"default:30" json:"reminder_minutes"`
	CreateGoogleMeet       bool      `gorm:"default:true" json:"create_google_meet"`
	MeetingLocation        string    `gorm:"size:255;default:'Google Meet'" json:"meeting_location"`
	UpdatedAt              time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (BookingSettings) TableName() string {
	return "booking_settings"
}

type InvoiceSettings struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	BizName       string    `gorm:"size:200;default:'JO4 Dev'" json:"biz_name"`
	BizAddress    string    `gorm:"size:400;default:''" json:"biz_address"`
	BizPhone      string    `gorm:"size:50;default:''" json:"biz_phone"`
	BizEmail      string    `gorm:"size:120;default:'jroetscyber@gmail.com'" json:"biz_email"`
	BankName      string    `gorm:"size:200;default:''" json:"bank_name"`
	AccountHolder string    `gorm:"size:200;default:''" json:"account_holder"`
	AccountNumber string    `gorm:"size:100;default:''" json:"account_number"`
	BranchCode    string    `gorm:"size:50;default:''" json:"branch_code"`
	VATNumber     string    `gorm:"size:50;default:''" json:"vat_number"`
	PaymentTerms  string    `gorm:"type:text;default:''" json:"payment_terms"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (InvoiceSettings) TableName() string {
	return "invoice_settings"
}

type AutomationLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ScriptName string    `gorm:"size:100;not null" json:"script_name"`
	Status     string    `gorm:"size:50;default:'COMPLETE'" json:"status"`
	Timestamp  time.Time `gorm:"autoCreateTime" json:"timestamp"`
}

func (AutomationLog) TableName() string {
	return "automation_logs"
}

type Analytics struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PagePath  string    `gorm:"size:255" json:"page_path"`
	VisitorIP string    `gorm:"size:50" json:"visitor_ip"`
	Timestamp time.Time `gorm:"autoCreateTime" json:"timestamp"`
}

func (Analytics) TableName() string {
	return "analytics"
}

type User struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	Username     string `gorm:"size:50;uniqueIndex;not null" json:"username"`
	PasswordHash string `gorm:"size:255;not null" json:"password_hash"`
}

func (User) TableName() string {
	return "login"
}

type Transaction struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Type        string    `gorm:"size:20;not null" json:"type"` // 'Income' or 'Expense'
	Category    string    `gorm:"size:100" json:"category"`
	Amount      float64   `gorm:"not null" json:"amount"`
	Description string    `gorm:"type:text" json:"description"`
	Date        time.Time `json:"date"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (Transaction) TableName() string {
	return "transactions"
}

type FAQSubmission struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Name       string    `gorm:"size:100;not null" json:"name"`
	Email      string    `gorm:"size:100;not null" json:"email"`
	Phone      string    `gorm:"size:50" json:"phone"`
	Question   string    `gorm:"type:text;not null" json:"question"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	IsAnswered bool      `gorm:"default:false" json:"is_answered"`
}

func (FAQSubmission) TableName() string {
	return "faq_submissions"
}

type Order struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	SequenceNumber int        `gorm:"index;not null" json:"sequence_number"` // Daily or sequential callout number: 101, 102...
	TrackingCode   string     `gorm:"size:64;uniqueIndex;not null" json:"tracking_code"` // Secure slug: e.g. "ord_a7f92b"
	CustomerName   string     `gorm:"size:100;not null" json:"customer_name"`
	CustomerPhone  string     `gorm:"size:50" json:"customer_phone"`
	CustomerEmail  string     `gorm:"size:100" json:"customer_email"`
	Items          string     `gorm:"type:text;not null" json:"items"`
	TotalAmount    float64    `gorm:"default:0" json:"total_amount"`
	Status         string     `gorm:"size:30;default:'received';index" json:"status"` // 'received', 'in_progress', 'ready', 'completed', 'cancelled'
	Notes          string     `gorm:"type:text" json:"notes"`
	NotifyCount    int        `gorm:"default:0" json:"notify_count"`
	ReadyAt        *time.Time `json:"ready_at"`
	CompletedAt    *time.Time `json:"completed_at"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Order) TableName() string {
	return "orders"
}
