package models

import (
	"time"
)

// Currency represents supported currencies
type Currency string

const (
	USD Currency = "USD"
	GEL Currency = "GEL"
)

// BillStatus represents the current state of a bill
type BillStatus string

const (
	StatusOpen   BillStatus = "OPEN"
	StatusClosed BillStatus = "CLOSED"
)

type BillingPeriodWorkflow struct {
	WorkflowID        string    `json:"bill_id"`
	BillingPeriodDays int       `json:"billing_period_days"`
	StartDate         time.Time `json:"start_date"`
	CustomerID        string    `json:"customer_id"`
	Bills             []*Bill   `json:"bills"`
}

// Bill represents a billing period for a customer
type Bill struct {
	ID          string      `json:"id"`
	Status      BillStatus  `json:"status"`
	Currency    Currency    `json:"currency"`
	TotalAmount float64     `json:"total_amount"`
	LineItems   []*LineItem `json:"line_items"`
	CreatedAt   time.Time   `json:"created_at"`
	ClosedAt    time.Time   `json:"closed_at,omitempty"`
	CloseReason string      `json:"close_reason,omitempty"`
	WorkflowID  string      `json:"workflow_id,omitempty"`
}

// LineItem represents a charge or fee within a bill
type LineItem struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	Quantity    int       `json:"quantity"`
	AddedAt     time.Time `json:"added_at"`
}

type StartBillingPeriodRequest struct {
	CustomerID        string   `json:"customer_id"`
	Currency          Currency `json:"currency"`
	BillingPeriodDays int      `json:"billing_period_days"`
}

// CreateBillRequest represents the request to create a new bill
type CreateBillRequest struct {
	CustomerID string `json:"customer_id"`
	Currency   string `json:"currency"`
}

// CreateBillResponse represents the response when creating a bill
type CreateBillResponse struct {
	BillID     string `json:"bill_id"`
	WorkflowID string `json:"workflow_id"`
}

// AddLineItemRequest represents the request to add a line item
type AddLineItemRequest struct {
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Quantity    int     `json:"quantity"`
	Currency    string  `json:"currency"`
}

// AddLineItemResponse represents the response when adding a line item
type AddLineItemResponse struct {
	LineItem *LineItem `json:"line_item"`
	Bill     *Bill     `json:"bill"`
}

// CloseBillRequest represents the request to close a bill
type CloseBillRequest struct {
	Reason string `json:"reason"`
}

// CloseBillResponse represents the response when closing a bill
type CloseBillResponse struct {
	Bill        *Bill     `json:"bill"`
	TotalAmount float64   `json:"total_amount"`
	TotalItems  int       `json:"total_items"`
	ClosedAt    time.Time `json:"closed_at"`
}

// ListBillsRequest represents query parameters for listing bills
type ListBillsRequest struct {
	Status string `json:"status,omitempty"`
}

// ListBillsResponse represents the response when listing bills
type ListBillsResponse struct {
	Bills []*Bill `json:"bills"`
	Total int64   `json:"total"`
}

type GetBillRequest struct {
	CustomerID string `json:"customer_id"`
	BillID     string `json:"bill_id"`
}

// GetBillResponse represents the response when getting a single bill
type GetBillResponse struct {
	Bill Bill `json:"bill"`
}

type CloseBillingPeriodResponse struct {
	WorkflowID     string  `json:"workflow_id"`
	Bills          []*Bill `json:"bills"`
	FinalAmountUSD float64 `json:"final_amount_usd"`
	FinalAmountGEL float64 `json:"final_amount_gel"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service,omitempty"`
	TaskQueue string `json:"task_queue,omitempty"`
	Error     string `json:"error,omitempty"`
}

// Validation methods

// IsValidCurrency checks if the currency is supported
func (c Currency) IsValid() bool {
	return c == USD || c == GEL
}

// IsValidStatus checks if the bill status is valid
func (s BillStatus) IsValid() bool {
	return s == StatusOpen || s == StatusClosed
}

// CanAddLineItems returns true if line items can be added to this bill
func (b *Bill) CanAddLineItems() bool {
	return b.Status == StatusOpen
}

// CalculateTotal calculates the total amount from all line items
func (b *Bill) CalculateTotal() float64 {
	total := 0.0
	for _, item := range b.LineItems {
		itemTotal := item.Amount * float64(item.Quantity)
		total += itemTotal
	}
	return total
}

// AddLineItem adds a new line item to the bill and updates the total
func (b *Bill) AddLineItem(lineItem *LineItem) {
	b.LineItems = append(b.LineItems, lineItem)
	b.TotalAmount += lineItem.Amount * float64(lineItem.Quantity)
}

// Close closes the bill with the given reason
func (b *Bill) Close(reason string) {
	now := time.Now()
	b.Status = StatusClosed
	b.ClosedAt = now
	b.CloseReason = reason
	b.TotalAmount = b.CalculateTotal()
}

func ConvertCurrencyAmount(currency1 Currency, currency2 Currency, amount float64) float64 {
	// Placeholder for currency conversion logic
	// In a real implementation, this would call an external service or use a conversion table
	if currency1 == currency2 {
		return amount
	}
	if currency1 == USD && currency2 == GEL {
		return amount * 2.5 // Example conversion rate
	}
	if currency1 == GEL && currency2 == USD {
		return amount * 0.4 // Example conversion rate
	}
	return amount // Default fallback
}
