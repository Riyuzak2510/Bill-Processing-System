package workflows

import (
	"testing"
	"time"

	"encore.app/models"
	"github.com/stretchr/testify/assert"
)

func TestFindBillState(t *testing.T) {
	t.Parallel()
	bills := []*models.Bill{
		{ID: "1"},
		{ID: "2"},
	}
	t.Run("found", func(t *testing.T) {
		bill := FindBillState(bills, "2")
		assert.NotNil(t, bill)
		assert.Equal(t, "2", bill.ID)
	})
	t.Run("not found", func(t *testing.T) {
		bill := FindBillState(bills, "3")
		assert.Nil(t, bill)
	})
}

func TestGetAccrualFactor(t *testing.T) {
	t.Run("returns default factor for recent startTime", func(t *testing.T) {
		startTime := time.Now()
		currentTime := time.Now()
		factor := getAccrualFactor(startTime, currentTime)
		assert.Equal(t, 1.0, factor)
	})

	t.Run("returns 2.5 for startTime older than 24 hours", func(t *testing.T) {
		startTime := time.Now().Add(-25 * time.Hour)
		currentTime := time.Now()
		factor := getAccrualFactor(startTime, currentTime)
		assert.Equal(t, 2.5, factor)
	})

	t.Run("returns 1.0 for startTime exactly 24 hours ago", func(t *testing.T) {
		startTime := time.Now().Add(-24 * time.Hour)
		currentTime := time.Now()
		factor := getAccrualFactor(startTime, currentTime)
		assert.Equal(t, 1.0, factor)
	})
}
