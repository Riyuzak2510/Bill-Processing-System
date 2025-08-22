package bills

import (
	"context"
	"fmt"
	"time"

	"encore.dev/rlog"
	"github.com/google/uuid"
	"go.temporal.io/sdk/client"

	"encore.app/constants"
	"encore.app/models"
	"encore.app/workflows"
)

//encore:api public method=POST path=/bills/startbillingperiod
func StartBillingPeriod(ctx context.Context, req *models.StartBillingPeriodRequest) error {
	// Validate request
	if err := validateStartBillingPeriodRequest(req); err != nil {
		return err
	}
	startTime := time.Now()

	workflowID := fmt.Sprintf("billing-period-workflow-%s-%s", startTime.Format("20060102"), req.CustomerID)

	// Start Temporal workflow
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: service.GetTaskQueue(),
	}

	workflowInput := &models.BillWorkflowInput{
		WorkflowID:        workflowID,
		CustomerID:        req.CustomerID,
		Currency:          models.Currency(req.Currency),
		BillingPeriodDays: req.BillingPeriodDays,
		StartedAt:         startTime,
		BillStates:        []*models.Bill{},
	}

	workflowRun, err := service.GetTemporalClient().ExecuteWorkflow(
		ctx, workflowOptions, workflows.BillWorkflow, workflowInput,
	)
	if err != nil {
		return fmt.Errorf("failed to start bill workflow: %w", err)
	}
	rlog.Info("started billing period workflow",
		"workflow_id", workflowRun.GetID(),
		"run_id", workflowRun.GetRunID(),
	)
	service.workflows[req.CustomerID] = workflowRun.GetID()

	return nil
}

//encore:api public method=POST path=/bills/createbill
func CreateBill(ctx context.Context, req *models.CreateBillRequest) (*models.CreateBillResponse, error) {
	// Validate request
	if err := validateCreateBillRequest(req); err != nil {
		return nil, err
	}
	workflowID, customerBillingPeriodFound := service.GetWorkflowIDForCustomer(req.CustomerID)
	if !customerBillingPeriodFound {
		return nil, fmt.Errorf("billing period not started for customer %s", req.CustomerID)
	}

	// Generate unique IDs
	billID := uuid.New().String()

	// Create bill
	signalInput := &models.CreateBillSignal{
		BillID:     billID,
		Currency:   models.Currency(req.Currency),
		WorkflowID: workflowID,
	}

	// Store bill
	err := service.temporalClient.SignalWorkflow(ctx, workflowID, "", constants.CreateBillSignalName, signalInput)
	if err != nil {
		return nil, fmt.Errorf("failed to signal workflow: %w", err)
	}

	return &models.CreateBillResponse{
		BillID:     billID,
		WorkflowID: workflowID,
	}, nil
}

//encore:api public method=POST path=/bills/addItem/:customerId/:billId
func AddLineItem(ctx context.Context, customerId string, billId string, req *models.AddLineItemRequest) (*models.AddLineItemResponse, error) {
	// Validate request
	if err := validateAddLineItemRequest(req); err != nil {
		return nil, err
	}
	workflowId, found := service.GetWorkflowIDForCustomer(customerId)
	if !found {
		return nil, fmt.Errorf("workflow not found")
	}
	// Get bill
	bill, err := getBillByID(ctx, billId, workflowId)
	if err != nil {
		return nil, fmt.Errorf("bill not found: %w", err)
	}

	// Check if bill can accept line items
	if !bill.CanAddLineItems() {
		return nil, fmt.Errorf("cannot add line items to closed bill")
	}

	// Create line item
	lineItem := models.LineItem{
		ID:          uuid.New().String(),
		Description: req.Description,
		Amount:      req.Amount,
		Quantity:    req.Quantity,
		AddedAt:     time.Now(),
	}

	// Send signal to workflow
	signalInput := models.AddLineItemSignal{
		LineItem: &lineItem,
		BillID:   bill.ID,
		Currency: models.Currency(req.Currency),
	}

	err = service.GetTemporalClient().SignalWorkflow(
		ctx, workflowId, "", constants.AddLineItemSignalName, signalInput,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to signal workflow: %w", err)
	}

	rlog.Info("added line item to bill",
		"bill_id", bill.ID,
		"line_item_id", lineItem.ID,
		"amount", lineItem.Amount,
	)

	return &models.AddLineItemResponse{
		LineItem: &lineItem,
		Bill:     bill,
	}, nil
}

//encore:api public method=POST path=/bills/close/:customerId/:billId
func CloseBill(ctx context.Context, customerId string, billId string, req *models.CloseBillRequest) (*models.CloseBillResponse, error) {
	// Get bill
	workflowId, found := service.GetWorkflowIDForCustomer(customerId)
	if !found {
		return nil, fmt.Errorf("workflow not found")
	}
	bill, err := getBillByID(ctx, billId, workflowId)
	if err != nil {
		return nil, fmt.Errorf("bill not found: %w", err)
	}

	// Check if bill is already closed
	if bill.Status == models.StatusClosed {
		return nil, fmt.Errorf("bill is already closed")
	}

	// Send signal to workflow
	signalInput := models.CloseBillSignal{
		Reason: req.Reason,
		BillID: billId,
	}

	err = service.GetTemporalClient().SignalWorkflow(
		ctx, workflowId, "", constants.CloseBillSignalName, signalInput,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to signal workflow: %w", err)
	}

	rlog.Info("closed bill",
		"bill_id", billId,
		"reason", req.Reason,
		"total_amount", bill.TotalAmount,
	)

	bill, err = getBillByID(ctx, billId, workflowId)
	if err != nil {
		return nil, fmt.Errorf("bill not found: %w", err)
	}

	return &models.CloseBillResponse{
		Bill:        bill,
		TotalAmount: bill.TotalAmount,
		TotalItems:  len(bill.LineItems),
		ClosedAt:    bill.ClosedAt,
	}, nil
}

//encore:api public method=POST path=/bills/getBill/:customerId/:billId
func GetBill(ctx context.Context, customerId string, billId string) (*models.GetBillResponse, error) {
	workflowId, found := service.GetWorkflowIDForCustomer(customerId)
	if !found {
		return nil, fmt.Errorf("workflow not found")
	}
	bill, err := getBillByID(ctx, billId, workflowId)
	if err != nil {
		return nil, fmt.Errorf("bill not found: %w", err)
	}

	return &models.GetBillResponse{
		Bill: *bill,
	}, nil
}

//encore:api public method=POST path=/bills/listBills/:customerId
func ListBills(ctx context.Context, customerId string, req *models.ListBillsRequest) (*models.ListBillsResponse, error) {
	// Parse query parameters from context
	// Note: In a real implementation, you'd extract these from query parameters

	workflowId, found := service.GetWorkflowIDForCustomer(customerId)
	if !found {
		return nil, fmt.Errorf("workflow not found for customer %s", customerId)
	}

	if err := validateListBillsRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	queryResult, err := service.GetTemporalClient().QueryWorkflow(ctx, workflowId, "", constants.ListBillsQuery, *req)
	if err != nil {
		return nil, fmt.Errorf("failed to query workflow: %w", err)
	}
	var result []*models.Bill

	err = queryResult.Get(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to get query result: %w", err)
	}

	return &models.ListBillsResponse{
		Bills: result,
		Total: int64(len(result)),
	}, nil
}

//encore:api public method=GET path=/bills/health
func HealthCheck(ctx context.Context) (*models.HealthResponse, error) {
	// Check if service is initialized
	if service == nil {
		return &models.HealthResponse{
			Status: "unhealthy",
			Error:  "Service not initialized",
		}, nil
	}

	// Check Temporal client connection
	_, err := service.GetTemporalClient().CheckHealth(ctx, &client.CheckHealthRequest{})
	if err != nil {
		return &models.HealthResponse{
			Status: "unhealthy",
			Error:  fmt.Sprintf("Temporal connection failed: %v", err),
		}, nil
	}

	return &models.HealthResponse{
		Status:    "healthy",
		Service:   "bills",
		TaskQueue: service.GetTaskQueue(),
	}, nil
}

//encore:api public method=POST path=/bills/closeBillingPeriod/:customerId
func CloseBillingPeriod(ctx context.Context, customerId string) (*models.CloseBillingPeriodResponse, error) {
	workflowId, found := service.GetWorkflowIDForCustomer(customerId)
	if !found {
		return nil, fmt.Errorf("workflow not found")
	}
	defer func() {
		cancelErr := service.temporalClient.CancelWorkflow(ctx, workflowId, "")
		if cancelErr != nil {
			rlog.Error("failed to cancel workflow", "error", cancelErr, "workflow_id", workflowId)
		}
		delete(service.workflows, customerId)
	}()

	err := service.GetTemporalClient().SignalWorkflow(ctx, workflowId, "", constants.CloseBillingPeriodSignalName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to signal workflow: %w", err)
	}

	var finalizedBills []*models.Bill
	queryResult, err := service.temporalClient.QueryWorkflow(ctx, workflowId, "", constants.ListBillsQuery, models.ListBillsRequest{
		Status: string(models.StatusClosed),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query finalized bills: %w", err)
	}
	err = queryResult.Get(&finalizedBills)

	if err != nil {
		return nil, fmt.Errorf("failed to get finalized bills from query result: %w", err)
	}

	finalBillAmountUSD := 0.0

	for _, bill := range finalizedBills {
		finalBillAmountUSD += models.ConvertCurrencyAmount(bill.Currency, models.USD, bill.TotalAmount)
	}

	finalAmountGEL := models.ConvertCurrencyAmount(models.USD, models.GEL, finalBillAmountUSD)

	return &models.CloseBillingPeriodResponse{
		WorkflowID:     workflowId,
		Bills:          finalizedBills,
		FinalAmountUSD: finalBillAmountUSD,
		FinalAmountGEL: finalAmountGEL,
	}, nil
}

func getBillByID(ctx context.Context, id string, workflowId string) (*models.Bill, error) {
	req := models.GetBillRequest{
		BillID: id,
	}
	queryResult, err := service.temporalClient.QueryWorkflow(ctx, workflowId, "", constants.GetBillQuery, req)
	if err != nil {
		return nil, fmt.Errorf("failed to query workflow: %w", err)
	}
	var bill *models.Bill
	if err := queryResult.Get(&bill); err != nil {
		return nil, fmt.Errorf("failed to get bill from query result: %w", err)
	}

	return bill, nil
}
