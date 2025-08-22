package models

import (
	"time"
)

// BillWorkflowInput represents the input for starting a bill workflow
type BillWorkflowInput struct {
	WorkflowID        string    `json:"workflow_id"`
	CustomerID        string    `json:"customer_id"`
	Currency          Currency  `json:"currency"`
	BillingPeriodDays int       `json:"billing_period_days"`
	BillStates        []*Bill   `json:"bill_states"`
	StartedAt         time.Time `json:"started_at"`
}

// AddLineItemSignal represents the signal to add a line item
type AddLineItemSignal struct {
	LineItem *LineItem `json:"line_item"`
	BillID   string    `json:"bill_id"`
	Currency Currency  `json:"currency"`
}

// CloseBillSignal represents the signal to close a bill
type CloseBillSignal struct {
	Reason string `json:"reason"`
	BillID string `json:"bill_id"`
}

type CreateBillSignal struct {
	Currency   Currency `json:"currency"`
	WorkflowID string   `json:"workflow_id"`
	BillID     string   `json:"bill_id"`
}
