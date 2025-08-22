package bills

import (
	"context"
	"testing"

	"encore.app/models" // Encore's test support package``
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateCloseWorkflow(t *testing.T) {
	t.Parallel()
	testCustomerId := uuid.New().String()
	nonExistentCustomerId := uuid.New().String()
	t.Run("Workflow Created And Closed", func(t *testing.T) {
		ctx := context.Background()
		err := StartBillingPeriod(ctx, &models.StartBillingPeriodRequest{
			CustomerID:        testCustomerId,
			Currency:          models.USD,
			BillingPeriodDays: 30,
		})
		defer CloseBillingPeriod(ctx, testCustomerId)
		assert.NoError(t, err)

		_, workflowFound := service.workflows[testCustomerId]
		assert.True(t, workflowFound)
		assert.Len(t, service.workflows, 1, "Expected one workflow to be created")
	})

	t.Run("Workflow Not Found", func(t *testing.T) {
		ctx := context.Background()
		_, err := CloseBillingPeriod(ctx, nonExistentCustomerId)
		assert.Error(t, err)
	})
}
func TestAddLineItemToBill(t *testing.T) {
	testCustomerId := uuid.New().String()
	nonExistentBillId := uuid.New().String()

	t.Run("Add Line Item", func(t *testing.T) {
		ctx := context.Background()

		err := StartBillingPeriod(ctx, &models.StartBillingPeriodRequest{
			CustomerID:        testCustomerId,
			Currency:          models.USD,
			BillingPeriodDays: 30,
		})
		defer CloseBillingPeriod(ctx, testCustomerId)
		assert.NoError(t, err)

		_, workflowFound := service.workflows[testCustomerId]
		assert.True(t, workflowFound)
		assert.Len(t, service.workflows, 1, "Expected one workflow to be created")

		// Create a bill
		createBillResp, err := CreateBill(ctx, &models.CreateBillRequest{
			CustomerID: testCustomerId,
			Currency:   string(models.USD),
		})
		assert.NoError(t, err)

		// Add a line item to the bill
		addLineItemResp, err := AddLineItem(ctx, testCustomerId, createBillResp.BillID, &models.AddLineItemRequest{
			Description: "Test Item 1",
			Amount:      100.0,
			Quantity:    1,
			Currency:    string(models.USD),
		})
		assert.NoError(t, err)
		assert.NotNil(t, addLineItemResp)

		addLineItemResp, err = AddLineItem(ctx, testCustomerId, createBillResp.BillID, &models.AddLineItemRequest{
			Description: "Test Item 2",
			Amount:      250.0,
			Quantity:    10,
			Currency:    string(models.GEL),
		})
		assert.NoError(t, err)

		billResp, err := GetBill(ctx, testCustomerId, createBillResp.BillID)
		assert.NoError(t, err)
		assert.NotNil(t, billResp)
		assert.Len(t, billResp.Bill.LineItems, 2)
		assert.Equal(t, billResp.Bill.LineItems[0].Description, "Test Item 1")
		assert.Equal(t, billResp.Bill.LineItems[0].Amount, 100.0)
		assert.Equal(t, billResp.Bill.LineItems[0].Quantity, 1)
		assert.Equal(t, billResp.Bill.LineItems[1].Description, "Test Item 2")
		assert.Equal(t, billResp.Bill.LineItems[1].Amount, 100.0)
		assert.Equal(t, billResp.Bill.LineItems[1].Quantity, 10)

	})
	t.Run("Add Line Item To Non-Existent Bill", func(t *testing.T) {
		ctx := context.Background()
		err := StartBillingPeriod(ctx, &models.StartBillingPeriodRequest{
			CustomerID:        testCustomerId,
			Currency:          models.USD,
			BillingPeriodDays: 30,
		})
		defer CloseBillingPeriod(ctx, testCustomerId)
		_, workflowFound := service.workflows[testCustomerId]
		assert.True(t, workflowFound)
		assert.Len(t, service.workflows, 1, "Expected one workflow to be created")

		// Try to add a line item to a non-existent bill
		addLineItemResp, err := AddLineItem(ctx, testCustomerId, nonExistentBillId, &models.AddLineItemRequest{
			Description: "Test Item",
			Amount:      100.0,
			Quantity:    1,
		})
		assert.Error(t, err)
		assert.Nil(t, addLineItemResp)

	})

	t.Run("Add Line Item Without Workflow", func(t *testing.T) {
		ctx := context.Background()

		// Try to add a line item without starting a billing period
		addLineItemResp, err := AddLineItem(ctx, testCustomerId, nonExistentBillId, &models.AddLineItemRequest{
			Description: "Test Item",
			Amount:      100.0,
			Quantity:    1,
		})
		assert.Error(t, err)
		assert.Nil(t, addLineItemResp)
	})
}

func TestCloseBillWorkflow(t *testing.T) {
	testCustomerId := uuid.New().String()
	nonExistentBillId := uuid.New().String()
	nonExistentCustomerId := uuid.New().String()
	t.Run("Close Bill Workflow", func(t *testing.T) {
		ctx := context.Background()

		err := StartBillingPeriod(ctx, &models.StartBillingPeriodRequest{
			CustomerID:        testCustomerId,
			Currency:          models.USD,
			BillingPeriodDays: 30,
		})
		defer CloseBillingPeriod(ctx, testCustomerId)
		assert.NoError(t, err)

		_, workflowFound := service.workflows[testCustomerId]
		assert.True(t, workflowFound)
		assert.Len(t, service.workflows, 1, "Expected one workflow to be created")

		// Create a bill
		createBillResp, err := CreateBill(ctx, &models.CreateBillRequest{
			CustomerID: testCustomerId,
			Currency:   string(models.USD),
		})
		assert.NoError(t, err)

		billResp, err := GetBill(ctx, testCustomerId, createBillResp.BillID)
		assert.NoError(t, err)
		assert.NotNil(t, billResp)
		assert.Equal(t, billResp.Bill.Status, models.StatusOpen)

		// Try to add a line item to the bill
		addLineItemResp, err := AddLineItem(ctx, testCustomerId, billResp.Bill.ID, &models.AddLineItemRequest{
			Description: "Test Item",
			Amount:      100.0,
			Quantity:    1,
			Currency:    string(models.USD),
		})

		assert.NoError(t, err)
		assert.NotNil(t, addLineItemResp)
		assert.Equal(t, addLineItemResp.LineItem.Description, "Test Item")
		assert.Equal(t, addLineItemResp.LineItem.Amount, 100.0)
		assert.Equal(t, addLineItemResp.LineItem.Quantity, 1)

		addLineItemResp, err = AddLineItem(ctx, testCustomerId, billResp.Bill.ID, &models.AddLineItemRequest{
			Description: "Test Item 2",
			Amount:      200.0,
			Quantity:    2,
			Currency:    string(models.GEL),
		})

		assert.NoError(t, err)
		assert.NotNil(t, addLineItemResp)
		assert.Equal(t, addLineItemResp.LineItem.Description, "Test Item 2")
		assert.Equal(t, addLineItemResp.LineItem.Amount, 200.0)
		assert.Equal(t, addLineItemResp.LineItem.Quantity, 2)

		// Close the bill
		closeBillResp, err := CloseBill(ctx, testCustomerId, billResp.Bill.ID, &models.CloseBillRequest{
			Reason: "No longer needed",
		})
		assert.NoError(t, err)

		// Verify the bill is closed
		assert.Equal(t, closeBillResp.Bill.Status, models.StatusClosed)
		assert.NoError(t, err)
		assert.NotNil(t, closeBillResp)
		assert.Equal(t, closeBillResp.Bill.Status, models.StatusClosed)
		assert.Equal(t, closeBillResp.Bill.CloseReason, "No longer needed")
		assert.Equal(t, closeBillResp.TotalAmount, 260.0)

	})

	t.Run("Close Bill Workflow - Non-existent Customer", func(t *testing.T) {
		ctx := context.Background()

		closeBillResp, err := CloseBill(ctx, nonExistentCustomerId, nonExistentBillId, &models.CloseBillRequest{
			Reason: "No longer needed",
		})
		assert.Error(t, err)
		assert.Nil(t, closeBillResp)
	})
}

func TestListBills(t *testing.T) {
	testCustomerId := uuid.New().String()
	t.Run("List Bills", func(t *testing.T) {
		ctx := context.Background()

		err := StartBillingPeriod(ctx, &models.StartBillingPeriodRequest{
			CustomerID:        testCustomerId,
			Currency:          models.USD,
			BillingPeriodDays: 30,
		})
		defer CloseBillingPeriod(ctx, testCustomerId)
		assert.NoError(t, err)

		_, workflowFound := service.workflows[testCustomerId]
		assert.True(t, workflowFound)
		assert.Len(t, service.workflows, 1, "Expected one workflow to be created")

		// Create a bill
		createBill1Resp, err := CreateBill(ctx, &models.CreateBillRequest{
			CustomerID: testCustomerId,
			Currency:   string(models.USD),
		})
		assert.NoError(t, err)

		createBill2Resp, err := CreateBill(ctx, &models.CreateBillRequest{
			CustomerID: testCustomerId,
			Currency:   string(models.GEL),
		})
		assert.NoError(t, err)

		listBillsResp, err := ListBills(ctx, testCustomerId, &models.ListBillsRequest{
			Status: string(models.StatusOpen),
		})
		assert.NoError(t, err)
		assert.NotNil(t, listBillsResp)
		assert.Len(t, listBillsResp.Bills, 2)
		assert.Equal(t, listBillsResp.Bills[0].ID, createBill1Resp.BillID)
		assert.Equal(t, listBillsResp.Bills[1].ID, createBill2Resp.BillID)

		CloseBillResp, err := CloseBill(ctx, testCustomerId, createBill1Resp.BillID, &models.CloseBillRequest{
			Reason: "No longer needed",
		})
		assert.NoError(t, err)
		assert.NotNil(t, CloseBillResp)
		assert.Equal(t, CloseBillResp.Bill.Status, models.StatusClosed)

		listBillsResp, err = ListBills(ctx, testCustomerId, &models.ListBillsRequest{
			Status: string(models.StatusClosed),
		})
		assert.NoError(t, err)
		assert.NotNil(t, listBillsResp)
		assert.Len(t, listBillsResp.Bills, 1)
		assert.Equal(t, listBillsResp.Bills[0].ID, CloseBillResp.Bill.ID)

	})

	t.Run("List Bills - Non-existent Customer", func(t *testing.T) {
		ctx := context.Background()

		listBillsResp, err := ListBills(ctx, testCustomerId, &models.ListBillsRequest{
			Status: string(models.StatusOpen),
		})
		assert.Error(t, err)
		assert.Nil(t, listBillsResp)
	})
}

func TestGetBill(t *testing.T) {
	testCustomerId := uuid.New().String()
	nonExistentBillId := uuid.New().String()
	nonExistentCustomerId := uuid.New().String()
	t.Run("Get Bill", func(t *testing.T) {
		ctx := context.Background()

		err := StartBillingPeriod(ctx, &models.StartBillingPeriodRequest{
			CustomerID:        testCustomerId,
			Currency:          models.USD,
			BillingPeriodDays: 30,
		})
		defer CloseBillingPeriod(ctx, testCustomerId)
		assert.NoError(t, err)

		_, workflowFound := service.workflows[testCustomerId]
		assert.True(t, workflowFound)
		assert.Len(t, service.workflows, 1, "Expected one workflow to be created")

		// Create a bill
		createBillResp, err := CreateBill(ctx, &models.CreateBillRequest{
			CustomerID: testCustomerId,
			Currency:   string(models.USD),
		})
		assert.NoError(t, err)

		billResp, err := GetBill(ctx, testCustomerId, createBillResp.BillID)
		assert.NoError(t, err)
		assert.NotNil(t, billResp)
		assert.Equal(t, billResp.Bill.ID, createBillResp.BillID)

	})

	t.Run("Get Bill - Non-existent Customer", func(t *testing.T) {
		ctx := context.Background()

		billResp, err := GetBill(ctx, nonExistentCustomerId, nonExistentBillId)
		assert.Error(t, err)
		assert.Nil(t, billResp)
	})

	t.Run("Get Bill - Non-existent Bill", func(t *testing.T) {
		ctx := context.Background()

		err := StartBillingPeriod(ctx, &models.StartBillingPeriodRequest{
			CustomerID:        testCustomerId,
			Currency:          models.USD,
			BillingPeriodDays: 30,
		})
		defer CloseBillingPeriod(ctx, testCustomerId)
		assert.NoError(t, err)

		billResp, err := GetBill(ctx, testCustomerId, nonExistentBillId)
		assert.Error(t, err)
		assert.Nil(t, billResp)
	})

	t.Run("Get Closed Bill", func(t *testing.T) {
		ctx := context.Background()

		err := StartBillingPeriod(ctx, &models.StartBillingPeriodRequest{
			CustomerID:        testCustomerId,
			Currency:          models.USD,
			BillingPeriodDays: 30,
		})
		defer CloseBillingPeriod(ctx, testCustomerId)
		assert.NoError(t, err)

		// Create a bill
		createBillResp, err := CreateBill(ctx, &models.CreateBillRequest{
			CustomerID: testCustomerId,
			Currency:   string(models.USD),
		})
		assert.NoError(t, err)

		CloseBillResp, err := CloseBill(ctx, testCustomerId, createBillResp.BillID, &models.CloseBillRequest{
			Reason: "No longer needed",
		})
		assert.NoError(t, err)
		assert.NotNil(t, CloseBillResp)
		assert.Equal(t, CloseBillResp.Bill.Status, models.StatusClosed)

		billResp, err := GetBill(ctx, testCustomerId, CloseBillResp.Bill.ID)
		assert.NoError(t, err)
		assert.NotNil(t, billResp)
		assert.Equal(t, billResp.Bill.ID, CloseBillResp.Bill.ID)

	})
}

func TestCloseBillingPeriod(t *testing.T) {
	testCustomerId := uuid.New().String()
	t.Run("Close Billing Period", func(t *testing.T) {
		ctx := context.Background()

		err := StartBillingPeriod(ctx, &models.StartBillingPeriodRequest{
			CustomerID:        testCustomerId,
			Currency:          models.USD,
			BillingPeriodDays: 30,
		})
		assert.NoError(t, err)

		createBillResp, err := CreateBill(ctx, &models.CreateBillRequest{
			CustomerID: testCustomerId,
			Currency:   string(models.USD),
		})
		assert.NoError(t, err)

		addItemResp, err := AddLineItem(ctx, testCustomerId, createBillResp.BillID, &models.AddLineItemRequest{
			Description: "Test Item1",
			Amount:      100.0,
			Quantity:    1,
			Currency:    string(models.USD),
		})
		assert.NoError(t, err)
		assert.NotNil(t, addItemResp)

		CloseBillPeriodResp, err := CloseBillingPeriod(ctx, testCustomerId)
		assert.NoError(t, err)
		assert.NotNil(t, CloseBillPeriodResp)

		assert.Len(t, CloseBillPeriodResp.Bills, 1)
		assert.Equal(t, CloseBillPeriodResp.Bills[0].ID, createBillResp.BillID)
		assert.Equal(t, CloseBillPeriodResp.Bills[0].Status, models.StatusClosed)
		assert.Equal(t, CloseBillPeriodResp.FinalAmountUSD, 100.0)
		assert.Equal(t, CloseBillPeriodResp.FinalAmountGEL, 250.0)

	})
}
