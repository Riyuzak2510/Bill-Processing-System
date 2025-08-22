package workflows

import (
	"fmt"
	"time"

	"encore.app/constants"
	"encore.app/models"

	"go.temporal.io/sdk/workflow"
)

// BillWorkflow manages the lifecycle of a bill
func BillWorkflow(ctx workflow.Context, input *models.BillWorkflowInput) error {
	logger := workflow.GetLogger(ctx)

	// Set up query handlers for workflow state
	if err := setGetBillQueryHandler(ctx, input); err != nil {
		logger.Error("Failed to set get bill query handler", "error", err)
		return err
	}
	if err := setListBillsQueryHandler(ctx, input); err != nil {
		logger.Error("Failed to set list bills query handler", "error", err)
		return err
	}

	// Set up signal channels
	addLineItemCh := workflow.GetSignalChannel(ctx, constants.AddLineItemSignalName)
	closeBillCh := workflow.GetSignalChannel(ctx, constants.CloseBillSignalName)
	createBillCh := workflow.GetSignalChannel(ctx, constants.CreateBillSignalName)
	closeBillingPeriodCh := workflow.GetSignalChannel(ctx, constants.CloseBillingPeriodSignalName)

	// Calculate billing period duration
	billingDuration := time.Duration(input.BillingPeriodDays) * 24 * time.Hour

	logger.Info("Bill workflow started",
		"customer_id", input.CustomerID,
		"billing_period", billingDuration,
	)

	// Start periodic fee accrual (simulate progressive billing)

	// Main workflow loop - listen for signals until billing period ends
	timerFired := false
	timerFuture := workflow.NewTimer(ctx, billingDuration)
	for !timerFired {
		selector := workflow.NewSelector(ctx)

		// Handle add line item signals
		selector.AddReceive(addLineItemCh, func(c workflow.ReceiveChannel, more bool) {
			var signal models.AddLineItemSignal
			c.Receive(ctx, &signal)
			handleAddLineItemSignal(ctx, input, signal)
		})

		// Handle close bill signals
		selector.AddReceive(closeBillCh, func(c workflow.ReceiveChannel, more bool) {
			var signal models.CloseBillSignal
			c.Receive(ctx, &signal)
			handleCloseBillSignal(ctx, input, signal)
		})

		selector.AddReceive(createBillCh, func(c workflow.ReceiveChannel, more bool) {
			var signal models.CreateBillSignal
			c.Receive(ctx, &signal)
			handleCreateBillSignal(ctx, input, signal)
		})

		selector.AddReceive(closeBillingPeriodCh, func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, nil)
			closeAllBillsDueToTimeout(ctx, input)
		})

		// Handle billing period timeout for all bills
		selector.AddFuture(timerFuture, func(f workflow.Future) {
			timerFired = true
			closeAllBillsDueToTimeout(ctx, input)
		})

		selector.Select(ctx)
	}

	// Finalize all bills at the end of the workflow
	for _, bill := range input.BillStates {
		logger.Info("Bill workflow completed successfully",
			"bill_id", bill.ID,
			"final_total", bill.TotalAmount,
			"total_items", len(bill.LineItems),
			"status", bill.Status,
		)
	}

	return nil
}

// setGetBillQueryHandler sets up the query handler for getting a single bill by ID
func setGetBillQueryHandler(ctx workflow.Context, workflowState *models.BillWorkflowInput) error {
	return workflow.SetQueryHandler(ctx, constants.GetBillQuery, func(req models.GetBillRequest) (*models.Bill, error) {
		for index := range workflowState.BillStates {
			if workflowState.BillStates[index].ID == req.BillID {
				return workflowState.BillStates[index], nil
			}
		}
		return nil, fmt.Errorf("bill not found")
	})
}

// setListBillsQueryHandler sets up the query handler for listing bills with optional filtering
func setListBillsQueryHandler(ctx workflow.Context, workflowState *models.BillWorkflowInput) error {
	return workflow.SetQueryHandler(ctx, constants.ListBillsQuery, func(req models.ListBillsRequest) ([]*models.Bill, error) {
		filteredBills := make([]*models.Bill, 0)
		for index := range workflowState.BillStates {
			if workflowState.BillStates[index].Status != models.BillStatus(req.Status) {
				continue
			}
			filteredBills = append(filteredBills, workflowState.BillStates[index])
		}
		return filteredBills, nil
	})
}

// handleCreateBillSignal processes a CreateBillSignal and adds a new bill to the workflow state
func handleCreateBillSignal(ctx workflow.Context, workflowState *models.BillWorkflowInput, signal models.CreateBillSignal) {
	logger := workflow.GetLogger(ctx)
	newBill := &models.Bill{
		ID:          signal.BillID,
		Status:      models.StatusOpen,
		Currency:    signal.Currency,
		TotalAmount: 0.0,
		LineItems:   []*models.LineItem{},
		CreatedAt:   workflow.Now(ctx),
		WorkflowID:  signal.WorkflowID,
	}
	workflowState.BillStates = append(workflowState.BillStates, newBill)
	logger.Info("New bill created",
		"bill_id", newBill.ID,
		"currency", newBill.Currency,
		"created_at", newBill.CreatedAt,
	)
}

// handleCloseBillSignal processes a CloseBillSignal and updates the bill state
func handleCloseBillSignal(ctx workflow.Context, workflowState *models.BillWorkflowInput, signal models.CloseBillSignal) {
	logger := workflow.GetLogger(ctx)
	billState := FindBillState(workflowState.BillStates, signal.BillID)
	if billState == nil {
		logger.Error("Bill not found for close signal",
			"bill_id", signal.BillID,
			"reason", signal.Reason,
		)
		return
	}
	billState.Close(signal.Reason)

	logger.Info("Bill closed via signal",
		"bill_id", billState.ID,
		"reason", signal.Reason,
		"final_total", billState.TotalAmount,
	)
}

// handleAddLineItemSignal processes an AddLineItemSignal and updates the bill state
func handleAddLineItemSignal(ctx workflow.Context, workflowState *models.BillWorkflowInput, signal models.AddLineItemSignal) {
	// Add timestamp to line item
	logger := workflow.GetLogger(ctx)
	signal.LineItem.AddedAt = workflow.Now(ctx)
	accrualFactor := getAccrualFactor(workflowState.StartedAt, signal.LineItem.AddedAt)
	billState := FindBillState(workflowState.BillStates, signal.BillID)
	signal.LineItem.Amount = models.ConvertCurrencyAmount(signal.Currency, billState.Currency, signal.LineItem.Amount) * accrualFactor
	// Add line item to bill state
	itemCopy := *signal.LineItem
	billState.LineItems = append(billState.LineItems, &itemCopy)
	billState.TotalAmount += signal.LineItem.Amount

	logger.Info("Line item added to bill",
		"bill_id", billState.ID,
		"item_id", signal.LineItem.ID,
		"amount", signal.LineItem.Amount,
		"new_total", billState.TotalAmount,
	)
}

func closeAllBillsDueToTimeout(ctx workflow.Context, input *models.BillWorkflowInput) {
	logger := workflow.GetLogger(ctx)

	for index := range input.BillStates {
		if input.BillStates[index].Status != models.StatusClosed {
			input.BillStates[index].Close("Billing period timed out")
			// Optionally recalculate total if needed
			// bill.TotalAmount = calculateTotal(bill.LineItems)
			logger.Info("Bill closed due to billing period completion",
				"bill_id", input.BillStates[index].ID,
				"final_total", input.BillStates[index].TotalAmount,
			)
		}
	}
}
