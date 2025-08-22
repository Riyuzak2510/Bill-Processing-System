package bills

import (
	"testing"

	"encore.app/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestValidateCreateBillRequest(t *testing.T) {
	testCustomerId := uuid.New().String()
	t.Parallel()
	t.Run("valid request", func(t *testing.T) {
		err := validateCreateBillRequest(&models.CreateBillRequest{
			CustomerID: testCustomerId,
			Currency:   string(models.USD),
		})
		assert.NoError(t, err)
	})
	t.Run("missing customer id", func(t *testing.T) {
		err := validateCreateBillRequest(&models.CreateBillRequest{
			Currency: string(models.USD),
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "customer_id is required")
	})
	t.Run("invalid currency", func(t *testing.T) {
		err := validateCreateBillRequest(&models.CreateBillRequest{
			CustomerID: testCustomerId,
			Currency:   "INVALID",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid currency")
	})
}

func TestValidateAddLineItemRequest(t *testing.T) {
	t.Parallel()
	t.Run("valid request", func(t *testing.T) {
		err := validateAddLineItemRequest(&models.AddLineItemRequest{
			Description: "desc",
			Amount:      10,
			Quantity:    1,
			Currency:    string(models.USD),
		})
		assert.NoError(t, err)
	})
	t.Run("missing description", func(t *testing.T) {
		err := validateAddLineItemRequest(&models.AddLineItemRequest{
			Amount:   10,
			Quantity: 1,
			Currency: string(models.USD),
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "description is required")
	})
	t.Run("negative amount", func(t *testing.T) {
		err := validateAddLineItemRequest(&models.AddLineItemRequest{
			Description: "desc",
			Amount:      -1,
			Quantity:    1,
			Currency:    string(models.USD),
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "amount must be non-negative")
	})
	t.Run("zero quantity", func(t *testing.T) {
		err := validateAddLineItemRequest(&models.AddLineItemRequest{
			Description: "desc",
			Amount:      1,
			Quantity:    0,
			Currency:    string(models.USD),
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "quantity must be positive")
	})
	t.Run("invalid currency", func(t *testing.T) {
		err := validateAddLineItemRequest(&models.AddLineItemRequest{
			Description: "desc",
			Amount:      10,
			Quantity:    1,
			Currency:    "INVALID",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid currency")
	})
}

func TestValidateStartBillingPeriodRequest(t *testing.T) {
	testCustomerId := uuid.New().String()
	t.Parallel()
	t.Run("valid request", func(t *testing.T) {
		err := validateStartBillingPeriodRequest(&models.StartBillingPeriodRequest{
			CustomerID:        testCustomerId,
			BillingPeriodDays: 10,
		})
		assert.NoError(t, err)
	})
	t.Run("missing customer id", func(t *testing.T) {
		err := validateStartBillingPeriodRequest(&models.StartBillingPeriodRequest{
			BillingPeriodDays: 10,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "customer_id is required")
	})
	t.Run("non-positive billing period", func(t *testing.T) {
		err := validateStartBillingPeriodRequest(&models.StartBillingPeriodRequest{
			CustomerID:        testCustomerId,
			BillingPeriodDays: 0,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "billing_period_days must be positive")
	})
}

func TestValidateListBillsRequest(t *testing.T) {
	t.Parallel()
	t.Run("valid status", func(t *testing.T) {
		err := validateListBillsRequest(&models.ListBillsRequest{Status: "OPEN"})
		assert.NoError(t, err)
		err = validateListBillsRequest(&models.ListBillsRequest{Status: "CLOSED"})
		assert.NoError(t, err)
	})
	t.Run("invalid status", func(t *testing.T) {
		err := validateListBillsRequest(&models.ListBillsRequest{Status: "INVALID"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid status")
	})
}
