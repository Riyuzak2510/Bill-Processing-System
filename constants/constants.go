package constants

const (
	// AddLineItemSignalName is used to add line items to an open bill
	AddLineItemSignalName = "add-line-item"

	// CloseBillSignalName is used to close a bill before the billing period ends
	CloseBillSignalName = "close-bill"

	// UpdateBillSignalName is used to update bill metadata
	UpdateBillSignalName = "update-bill"

	// SuspendBillSignalName is used to suspend billing temporarily
	SuspendBillSignalName = "suspend-bill"

	// ResumeBillSignalName is used to resume suspended billing
	ResumeBillSignalName = "resume-bill"

	// CreateBillSignalName is used to create a new bill
	CreateBillSignalName = "create-bill"

	// CloseBillingPeriodSignalName is used to close a billing period
	CloseBillingPeriodSignalName = "close-billing-period"

	// GetBillQuery is used to retrieve a bill by ID
	GetBillQuery = "get-bill"

	// ListBillsQuery is used to retrieve all bills
	ListBillsQuery = "list-bills"
)
