package workflows

import (
	"testing"
	"time"

	"encore.app/constants"
	"encore.app/models"

	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/testsuite"
)

type BillWorkflowTestSuite struct {
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
}

func TestBillWorkflowTestSuite(t *testing.T) {
	suite := new(BillWorkflowTestSuite)
	suite.env = suite.NewTestWorkflowEnvironment()
	defer suite.env.AssertExpectations(t)

	// Register the workflow under test
	suite.env.RegisterWorkflow(BillWorkflow)

	// Run individual test cases
	t.Run("Create, Add Items, Close", suite.TestBillWorkflowLifecycle)

	suite.env = suite.NewTestWorkflowEnvironment()
	t.Run("Create, Add Items, Close with Timer", suite.TestBillWorkflowLifecycleTimer)
}

func (s *BillWorkflowTestSuite) TestBillWorkflowLifecycle(t *testing.T) {
	// 1) Seed clock so workflow.Now() is deterministic
	start := time.Date(2025, 8, 21, 7, 0, 0, 0, time.UTC)
	s.env.SetStartTime(start)

	// 2) Prepare initial input
	input := &models.BillWorkflowInput{
		WorkflowID:        "wf-1",
		CustomerID:        "cust-1",
		Currency:          "USD",
		BillingPeriodDays: 1, // 1 day for quick timer
		BillStates:        []*models.Bill{{ID: "bill-1", Status: models.StatusOpen}},
	}

	// 3) Schedule signals:
	//   a) After 1 minute, create a bill
	s.env.RegisterDelayedCallback(func() {
		res, err := s.env.QueryWorkflow(constants.ListBillsQuery, &models.ListBillsRequest{
			Status: string(models.StatusOpen),
		})
		var billStates []*models.Bill
		assert.NoError(t, err)
		err = res.Get(&billStates)
		assert.Equal(t, 1, len(billStates), "Expected one bill before creation")

		// Now create the bill
		s.env.SignalWorkflow(constants.CreateBillSignalName, models.CreateBillSignal{
			Currency:   models.USD,
			WorkflowID: "wf-1",
			BillID:     "bill-2",
		})
	}, time.Second)

	s.env.RegisterDelayedCallback(func() {
		res, err := s.env.QueryWorkflow(constants.ListBillsQuery, &models.ListBillsRequest{
			Status: string(models.StatusOpen),
		})
		var billStates []*models.Bill
		assert.NoError(t, err)
		err = res.Get(&billStates)
		assert.Equal(t, 2, len(billStates), "Expected two bills after creation")
	}, 2*time.Second)

	//   b) After 2 minutes, add a line item
	s.env.RegisterDelayedCallback(func() {

		newItem := &models.LineItem{
			ID:     "item-xyz",
			Amount: 100.0,
		}
		s.env.SignalWorkflow(constants.AddLineItemSignalName, &models.AddLineItemSignal{
			BillID:   "bill-1",
			LineItem: newItem,
		})
	}, 4*time.Second)

	s.env.RegisterDelayedCallback(func() {
		res, err := s.env.QueryWorkflow(constants.GetBillQuery, &models.GetBillRequest{BillID: "bill-1"})
		var bill *models.Bill
		err = res.Get(&bill)
		assert.NoError(t, err)
		assert.Equal(t, "bill-1", bill.ID)
		assert.Len(t, bill.LineItems, 1, "Expected one line item")
	}, 5*time.Second)

	s.env.RegisterDelayedCallback(func() {
		res, err := s.env.QueryWorkflow(constants.GetBillQuery, &models.GetBillRequest{BillID: "bill-1"})
		var bill *models.Bill
		err = res.Get(&bill)
		assert.NoError(t, err)
		s.env.SignalWorkflow(constants.CloseBillSignalName, models.CloseBillSignal{
			BillID: "bill-1",
			Reason: "Timer fired",
		})
	}, 15*time.Second)

	// 4) Execute the workflow
	s.env.ExecuteWorkflow(BillWorkflow, input)

	// 5) Assertions
	assert.True(t, s.env.IsWorkflowCompleted())
	assert.NoError(t, s.env.GetWorkflowError())
}

func (s *BillWorkflowTestSuite) TestBillWorkflowLifecycleTimer(t *testing.T) {
	// 1) Seed clock so workflow.Now() is deterministic
	start := time.Date(2025, 8, 21, 7, 0, 0, 0, time.UTC)
	s.env.SetStartTime(start)

	// 2) Prepare initial input
	input := &models.BillWorkflowInput{
		WorkflowID:        "wf-1",
		CustomerID:        "cust-1",
		Currency:          "USD",
		BillingPeriodDays: 1, // 1 day for quick timer
		BillStates:        []*models.Bill{{ID: "bill-1", Status: models.StatusOpen}},
	}

	// 3) Schedule signals:
	//   a) After 1 minute, create a bill
	s.env.RegisterDelayedCallback(func() {
		res, err := s.env.QueryWorkflow(constants.ListBillsQuery, &models.ListBillsRequest{
			Status: string(models.StatusOpen),
		})
		var billStates []*models.Bill
		assert.NoError(t, err)
		err = res.Get(&billStates)
		assert.Equal(t, 1, len(billStates), "Expected one bill before creation")

		// Now create the bill
		s.env.SignalWorkflow(constants.CreateBillSignalName, models.CreateBillSignal{
			Currency:   "USD",
			WorkflowID: "wf-1",
			BillID:     "bill-2",
		})
	}, time.Second)

	s.env.RegisterDelayedCallback(func() {
		res, err := s.env.QueryWorkflow(constants.ListBillsQuery, &models.ListBillsRequest{
			Status: string(models.StatusOpen),
		})
		var billStates []*models.Bill
		assert.NoError(t, err)
		err = res.Get(&billStates)
		assert.Equal(t, 2, len(billStates), "Expected two bills after creation")
	}, 2*time.Second)

	//   b) After 2 seconds, add a line item
	s.env.RegisterDelayedCallback(func() {

		newItem := &models.LineItem{
			ID:          "item-xyz",
			Amount:      100.0,
			Description: "XYZ",
			Quantity:    1,
		}
		s.env.SignalWorkflow(constants.AddLineItemSignalName, models.AddLineItemSignal{
			BillID:   "bill-1",
			LineItem: newItem,
		})
	}, 4*time.Second)

	s.env.RegisterDelayedCallback(func() {
		res, err := s.env.QueryWorkflow(constants.GetBillQuery, &models.GetBillRequest{BillID: "bill-1"})
		var bill *models.Bill
		err = res.Get(&bill)
		assert.NoError(t, err)
		assert.Equal(t, "bill-1", bill.ID)
		assert.Len(t, bill.LineItems, 1, "Expected one line item")
	}, 5*time.Second)

	s.env.RegisterDelayedCallback(func() {}, 24*time.Hour)

	// 4) Execute the workflow
	s.env.ExecuteWorkflow(BillWorkflow, input)

	// 5) Assertions
	res, err := s.env.QueryWorkflow(constants.ListBillsQuery, &models.ListBillsRequest{Status: string(models.StatusClosed)})
	var ClosedBillStates []*models.Bill
	assert.NoError(t, err)
	err = res.Get(&ClosedBillStates)
	assert.Len(t, ClosedBillStates, 2, "Expected two closed bills")
	res, err = s.env.QueryWorkflow(constants.ListBillsQuery, &models.ListBillsRequest{Status: string(models.StatusOpen)})
	var OpenBillStates []*models.Bill
	assert.NoError(t, err)
	err = res.Get(&OpenBillStates)
	assert.Len(t, OpenBillStates, 0, "Expected no open bills")
	assert.True(t, s.env.IsWorkflowCompleted())
	assert.NoError(t, s.env.GetWorkflowError())
}

func computeAccrualFactor(now, start time.Time) float64 {
	var factor float64 = 1.0
	if now.Sub(start) < 30*24*time.Hour {
		factor = 2.5
	}
	return factor
}

// exposed only for tests
func GetAccrualFactorForTest(now, start time.Time) float64 {
	return computeAccrualFactor(now, start)
}
