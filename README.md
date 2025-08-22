# Bill Processing System Documentation

## Overview
This project implements a bill processing system using Go, Encore, and Temporal. It exposes a set of RESTful APIs for bill lifecycle management and leverages Temporal workflows to orchestrate billing periods, bill creation, line item management, and closing operations in a reliable, fault-tolerant manner.

---

## Exposed APIs

### 1. Start Billing Period
- **Endpoint:** `POST /bills/startbillingperiod`
- **Description:** Starts a new billing period for a customer by launching a Temporal workflow.
- **Request Body:**
  - `customer_id` (string, required)
  - `currency` (string, required, e.g., "USD")
  - `billing_period_days` (int, required)
- **Response:** `200 OK` on success, error otherwise.

### 2. Create Bill
- **Endpoint:** `POST /bills/createbill`
- **Description:** Creates a new bill for a customer within an active billing period.
- **Request Body:**
  - `customer_id` (string, required)
  - `currency` (string, required)
- **Response:**
  - `bill_id` (string)
  - `workflow_id` (string)

### 3. Add Line Item
- **Endpoint:** `POST /bills/addItem/:customerId/:billId`
- **Description:** Adds a line item to a specific bill.
- **Request Body:**
  - `description` (string, required)
  - `amount` (float, required)
  - `quantity` (int, required)
  - `currency` (string, required)
- **Response:**
  - `line_item` (object)
  - `bill` (object)

### 4. Close Bill
- **Endpoint:** `POST /bills/close/:customerId/:billId`
- **Description:** Closes a specific bill, preventing further line items from being added.
- **Request Body:**
  - `reason` (string, required)
- **Response:**
  - `bill` (object)
  - `total_amount` (float)
  - `total_items` (int)
  - `closed_at` (timestamp)

### 5. Get Bill
- **Endpoint:** `GET /bills/getBill/:customerId/:billId`
- **Description:** Retrieves details of a specific bill.
- **Response:**
  - `bill` (object)

### 6. List Bills
- **Endpoint:** `POST /bills/listBills/:customerId`
- **Description:** Lists all bills for a customer, optionally filtered by status.
- **Request Body:**
  - `status` (string, optional: "OPEN" or "CLOSED")
- **Response:**
  - `bills` (array)
  - `total` (float)

### 7. Close Billing Period
- **Endpoint:** `POST /bills/closeBillingPeriod/:customerId`
- **Description:** Closes the billing period for a customer, closing all open bills.
- **Response:** `200 OK` on success.

### 8. Health Check
- **Endpoint:** `GET /bills/health`
- **Description:** Returns the health status of the service and Temporal connection.
- **Response:**
  - `status` (string)
  - `service` (string)
  - `task_queue` (string)
  - `error` (string, optional)

---

## Temporal Workflow Usage

### Why Temporal?
Temporal provides durable, reliable, and scalable workflow orchestration. In this project, it ensures that billing periods, bill creation, and closing operations are resilient to failures and can be managed over long durations.

### Workflow Design
- **BillWorkflow**: The core workflow manages the lifecycle of a billing period for a customer. It:
  - Sets up signal and query handlers for dynamic interaction (e.g., adding line items, closing bills, querying state).
  - Listens for signals to create bills, add line items, or close bills.
  - Waits for the billing period to elapse (using a timer) or for an explicit close signal.
  - On completion, closes all open bills and finalizes the billing period.

#### Key Features:
- **Signals**: Used to asynchronously add line items, create bills, or close bills during the workflow's execution.
- **Queries**: Allow external systems to query the current state of bills in real time.
- **Timers**: Ensure the workflow runs for the full billing period, automatically closing bills if not done manually.
- **Idempotency**: Workflow and API design ensure that repeated requests do not cause inconsistent state.

### Example Flow
1. `StartBillingPeriod` API starts a Temporal workflow for a customer.
2. `CreateBill` and `AddLineItem` APIs send signals to the workflow to mutate state.
3. `CloseBill` and `CloseBillingPeriod` APIs send signals to close bills or the entire period.
4. The workflow maintains all state in memory (durably persisted by Temporal) and exposes queries for real-time inspection.
5. When the billing period ends (timer fires) or is closed, the workflow finalizes all bills and completes.

---

## Error Handling
- All APIs return clear error messages and codes for invalid input, missing workflows, or Temporal failures.
- Validation helpers ensure only valid data is processed.

---

## Extensibility
- The system is designed for easy extension: new signals, queries, or workflow logic can be added with minimal changes.
- Currency conversion and accrual logic are pluggable for future enhancements.

---

## Accrual Factor Logic

### How Accrual Factor is Used
The accrual factor is used in the bill processing workflow to adjust the amount of a line item based on how much time has passed since the billing period started. When a new line item is added to a bill, the system calculates an accrual factor using the `getAccrualFactor` function. This factor is then multiplied with the (possibly converted) amount of the line item before it is added to the bill's total.

**Usage Example:**
When handling an AddLineItem signal in the workflow, the following logic is applied:

```go
accrualFactor := getAccrualFactor(workflowState.StartedAt)
signal.LineItem.Amount = convertCurrencyAmount(signal.Currency, billState.Currency, signal.LineItem.Amount) * accrualFactor
```

### Simple Time-Based Implementation
In this implementation, the accrual factor is determined by a simple rule:
- If the billing period started more than 24 hours ago, the accrual factor is set to 2.5.
- Otherwise, the accrual factor is 1.0 (no adjustment).

This is a placeholder for more complex business logic. In a real-world scenario, the accrual factor might depend on additional parameters, such as customer type, bill type, or external data sources. For demonstration purposes, we use this straightforward time-based approach to show how such logic can be integrated and easily extended in the future.

---

## Currency Support and Conversion

### Supported Currencies
This system currently supports only two types of currency:
- **USD** (United States Dollar)
- **GEL** (Georgian Lari)

### Currency Conversion Logic
When a line item is added to a bill, its amount is automatically converted to the currency of the bill, if necessary. The conversion is handled by the `convertCurrencyAmount` function, which uses a fixed conversion rate for demonstration purposes:
- USD to GEL: amount × 2.5
- GEL to USD: amount × 0.4
- If the currencies are the same, the amount is unchanged.

This ensures that all amounts within a bill are consistent and in the bill's original currency, regardless of the currency used when adding individual items. In a production system, this logic could be extended to support more currencies and dynamic exchange rates.

---

## Summary
This project demonstrates a robust, production-grade bill processing system using Encore and Temporal. It exposes a clean API surface for clients and leverages Temporal's workflow engine for reliability, observability, and operational simplicity.
