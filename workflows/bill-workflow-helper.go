package workflows

import (
	"time"

	"encore.app/models"
)

func FindBillState(billStates []*models.Bill, billID string) *models.Bill {
	for i := range billStates {
		if billStates[i].ID == billID {
			return billStates[i]
		}
	}
	return nil
}

func getAccrualFactor(startTime time.Time, currentTime time.Time) float64 {

	// Simulate fetching accrual factor
	// In a real implementation, this might involve complex calculations or external API calls
	accrualFactor := 1.0 // Default factor

	// Example logic: if the billing period started more than 24 hours ago, increase the factor
	if startTime.Before(currentTime.Add(-1 * 24 * time.Hour)) {
		accrualFactor = 2.5
	}

	return accrualFactor
}
