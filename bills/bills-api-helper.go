package bills

import (
	"fmt"

	"encore.app/models"
)

func validateCreateBillRequest(req *models.CreateBillRequest) error {
	if req.CustomerID == "" {
		return fmt.Errorf("customer_id is required")
	}
	if !models.Currency(req.Currency).IsValid() {
		return fmt.Errorf("invalid currency: %s (supported: USD, GEL)", req.Currency)
	}
	return nil
}

func validateAddLineItemRequest(req *models.AddLineItemRequest) error {
	if req.Description == "" {
		return fmt.Errorf("description is required")
	}
	if req.Amount < 0 {
		return fmt.Errorf("amount must be non-negative")
	}
	if req.Quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}
	if !models.Currency(req.Currency).IsValid() {
		return fmt.Errorf("invalid currency: %s (supported: USD, GEL)", req.Currency)
	}
	return nil
}

func validateStartBillingPeriodRequest(req *models.StartBillingPeriodRequest) error {
	if req.CustomerID == "" {
		return fmt.Errorf("customer_id is required")
	}
	if req.BillingPeriodDays <= 0 {
		return fmt.Errorf("billing_period_days must be positive")
	}
	return nil
}

func validateListBillsRequest(req *models.ListBillsRequest) error {
	if !models.BillStatus(req.Status).IsValid() {
		return fmt.Errorf("invalid status: %s (supported: OPEN, CLOSED)", req.Status)
	}
	return nil
}
