package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/JRoetscyber/my_website/go_app/internal/models"
	"gorm.io/gorm"
)

// GenerateTrackingCode creates a secure, URL-friendly tracking token
func GenerateTrackingCode() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("ord_%s", hex.EncodeToString(b))
}

// NextCalloutSequence returns the next sequential call-out number.
// Sequence starts at 101 each day (e.g. 101, 102, 103...) for clean counter callouts.
func NextCalloutSequence(db *gorm.DB) int {
	var lastOrder models.Order
	today := time.Now().Truncate(24 * time.Hour)
	err := db.Where("created_at >= ?", today).Order("sequence_number desc").First(&lastOrder).Error
	if err == nil && lastOrder.SequenceNumber >= 100 {
		return lastOrder.SequenceNumber + 1
	}
	return 101
}
