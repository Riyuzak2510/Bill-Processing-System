package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidEnums(t *testing.T) {
	t.Parallel()
	t.Run("Currency", func(t *testing.T) {
		t.Parallel()
		cases := []struct {
			name     string
			currency Currency
			want     bool
		}{
			{"USD valid", USD, true},
			{"GEL valid", GEL, true},
			{"EUR invalid", Currency("EUR"), false},
		}
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				assert.Equal(t, tc.want, tc.currency.IsValid())
			})
		}
	})
	t.Run("BillStatus", func(t *testing.T) {
		t.Parallel()
		cases := []struct {
			name   string
			status BillStatus
			want   bool
		}{
			{"Open valid", StatusOpen, true},
			{"Closed valid", StatusClosed, true},
			{"Invalid", BillStatus("INVALID"), false},
		}
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				assert.Equal(t, tc.want, tc.status.IsValid())
			})
		}
	})
}

func TestBill_CanAddLineItems(t *testing.T) {
	b := &Bill{Status: StatusOpen}
	assert.True(t, b.CanAddLineItems())
	b.Status = StatusClosed
	assert.False(t, b.CanAddLineItems())
}

func TestBill_CalculateTotal(t *testing.T) {
	b := &Bill{
		LineItems: []*LineItem{
			{Amount: 10, Quantity: 2},
			{Amount: 5, Quantity: 3},
		},
	}
	total := b.CalculateTotal()
	assert.Equal(t, 10.0*2+5.0*3, total)
}

func TestBill_AddLineItem(t *testing.T) {
	b := &Bill{LineItems: []*LineItem{}}
	item := &LineItem{Amount: 7, Quantity: 2}
	b.AddLineItem(item)
	assert.Len(t, b.LineItems, 1)
	assert.Equal(t, 14.0, b.TotalAmount)
}

func TestBill_Close(t *testing.T) {
	b := &Bill{
		Status:    StatusOpen,
		LineItems: []*LineItem{{Amount: 3, Quantity: 2}},
	}
	b.Close("done")
	assert.Equal(t, StatusClosed, b.Status)
	assert.NotZero(t, b.ClosedAt)
	assert.Equal(t, "done", b.CloseReason)
	assert.Equal(t, 6.0, b.TotalAmount)
}

func TestConvertCurrencyAmount(t *testing.T) {
	t.Parallel()
	t.Run("same currency", func(t *testing.T) {
		amt := ConvertCurrencyAmount(USD, USD, 10)
		assert.Equal(t, 10.0, amt)
	})
	t.Run("USD to GEL", func(t *testing.T) {
		amt := ConvertCurrencyAmount(USD, GEL, 10)
		assert.Equal(t, 25.0, amt)
	})
	t.Run("GEL to USD", func(t *testing.T) {
		amt := ConvertCurrencyAmount(GEL, USD, 10)
		assert.Equal(t, 4.0, amt)
	})
	t.Run("unsupported currency", func(t *testing.T) {
		amt := ConvertCurrencyAmount(Currency("EUR"), Currency("JPY"), 10)
		assert.Equal(t, 10.0, amt)
	})
}
