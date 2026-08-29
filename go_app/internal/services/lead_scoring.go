package services

import (
	"math"
	"strings"
	"time"
)

type LeadScoringInput struct {
	ClientCompany      string
	ContactRole        string
	Budget             float64
	ProjectType        string
	VisitedPages       []string
	WhatsappEngagement string
	PhoneNumber        string
	LastActivityDate   time.Time
}

type LeadScoreBreakdown struct {
	Explicit float64 `json:"explicit"`
	Implicit float64 `json:"implicit"`
	Urgency  float64 `json:"urgency"`
}

type LeadScoreResult struct {
	Score          float64            `json:"score"`
	Classification string             `json:"classification"`
	Breakdown      LeadScoreBreakdown `json:"breakdown"`
}

func CalculateLeadScore(input LeadScoringInput) LeadScoreResult {
	// 1. EXPLICIT DATA SCORING (Max 100)
	var explicitScore float64

	// Budget Scoring (33.3% of Explicit)
	var budgetPoints float64
	if input.Budget >= 500000 {
		budgetPoints = 100
	} else if input.Budget <= 1000 {
		budgetPoints = 10
	} else {
		budgetPoints = 10 + ((input.Budget - 1000) / (500000 - 1000) * 90)
	}
	explicitScore += budgetPoints * 0.333

	// Role Map (33.3% of Explicit)
	roleMap := map[string]float64{
		"owner":      100,
		"principal":  100,
		"director":   80,
		"manager":    50,
		"consultant": 20,
	}
	rolePoints, ok := roleMap[strings.ToLower(strings.TrimSpace(input.ContactRole))]
	if !ok {
		rolePoints = 10
	}
	explicitScore += rolePoints * 0.333

	// Project Type (33.4% of Explicit)
	projectMap := map[string]float64{
		"erp":             100,
		"portal":          100,
		"iot_integration": 90,
		"data_pipeline":   80,
		"crm":             70,
		"server_config":   70,
		"automation":      60,
		"pos_system":      60,
		"static":          20,
	}
	projectPoints, ok := projectMap[strings.ToLower(strings.TrimSpace(input.ProjectType))]
	if !ok {
		projectPoints = 10
	}
	explicitScore += projectPoints * 0.334

	// 2. IMPLICIT DATA SCORING (Max 100)
	var implicitScore float64

	// Website Analytics (40% of Implicit)
	highIntentPages := []string{"/projects", "/book", "/pricing", "case-study"}
	intentMatch := false
	for _, page := range input.VisitedPages {
		for _, high := range highIntentPages {
			if strings.Contains(page, high) {
				intentMatch = true
				break
			}
		}
		if intentMatch {
			break
		}
	}
	var webPoints float64 = 20
	if intentMatch {
		webPoints = 100
	}
	implicitScore += webPoints * 0.40

	// WhatsApp Engagement (40% of Implicit)
	waMap := map[string]float64{
		"replied": 100,
		"read":    50,
		"ignored": 0,
	}
	waPoints, ok := waMap[strings.ToLower(strings.TrimSpace(input.WhatsappEngagement))]
	if !ok {
		waPoints = 0
	}
	implicitScore += waPoints * 0.40

	// Data Completeness (20% of Implicit)
	var completeness float64
	if strings.TrimSpace(input.ClientCompany) != "" {
		completeness += 50
	}
	if strings.TrimSpace(input.PhoneNumber) != "" {
		completeness += 50
	}
	implicitScore += completeness * 0.20

	// 3. TIME DECAY / URGENCY (Max 100)
	var urgencyScore float64
	now := time.Now()
	lastAct := input.LastActivityDate
	if lastAct.IsZero() {
		lastAct = now
	}
	daysSince := int(now.Sub(lastAct).Hours() / 24)
	gracePeriod := 14

	if daysSince <= gracePeriod {
		urgencyScore = 100
	} else {
		k := 0.1
		t := float64(daysSince - gracePeriod)
		urgencyScore = 100 * math.Exp(-k*t)
	}

	// FINAL AGGREGATION
	finalScore := (explicitScore * 0.40) + (implicitScore * 0.40) + (urgencyScore * 0.20)
	finalScore = math.Round(finalScore*10) / 10

	classification := "Cold"
	if finalScore >= 80 {
		classification = "Hot"
	} else if finalScore >= 50 {
		classification = "Warm"
	}

	return LeadScoreResult{
		Score:          finalScore,
		Classification: classification,
		Breakdown: LeadScoreBreakdown{
			Explicit: math.Round(explicitScore*10) / 10,
			Implicit: math.Round(implicitScore*10) / 10,
			Urgency:  math.Round(urgencyScore*10) / 10,
		},
	}
}
