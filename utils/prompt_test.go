package utils

import (
	"testing"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/shopspring/decimal"
)

func getTestHoldings() []alpaca.Position {
	qtyAvailable := decimal.NewFromInt(34)
	marketValue := decimal.NewFromInt(2999)
	currentPrice := decimal.NewFromInt(87)
	lastdayPrice := decimal.NewFromInt(87)
	changeToday := decimal.NewFromInt(0)
	unrealizedPL := decimal.NewFromInt(0)
	return []alpaca.Position{
		{
			Symbol:       "ACGL",
			QtyAvailable: qtyAvailable,
			MarketValue:  &marketValue,
			CurrentPrice: &currentPrice,
			LastdayPrice: &lastdayPrice,
			ChangeToday:  &changeToday,
			UnrealizedPL: &unrealizedPL,
		},
	}
}

func TestPrepHoldingsString(t *testing.T) {
	holdings := getTestHoldings()

	got := prepHoldingsString(holdings)
	want := "1. Symbol: ACGL\nQuantity: 34\nMarket Value: 2999\nCurrent Price: 87\nLast Day Price: 87\nChange Today % (0-1): 0\nUnrealized PL: 0\nCost Basis: 0\n\n"
	if got != want {
		t.Errorf("prepHoldingsString() = %q, want %q", got, want)
	}
}
