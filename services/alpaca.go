package services

import (
	"fmt"
	"os"
	"time"

	a "github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/dickeyy/cis-320/types"
	"github.com/shopspring/decimal"
)

var (
	Alpaca        *a.Client  = nil
	AlpacaAccount *a.Account = nil
	devMode                  = false
)

func SetDevMode(on bool) { devMode = on }

func InitializeAlpaca() error {
	Alpaca = a.NewClient(a.ClientOpts{
		APIKey:    os.Getenv("ALPACA_KEY"),
		APISecret: os.Getenv("ALPACA_SECRET"),
		BaseURL:   os.Getenv("ALPACA_API"),
	})
	acct, err := Alpaca.GetAccount()
	if err != nil {
		return err
	}
	AlpacaAccount = acct
	return nil
}

func PlaceOrder(trade *types.Trade) (*types.Trade, error) {
	if devMode {
		// Simulate immediate fill in dev mode
		price := decimal.NewFromFloat(100.00)
		// If Amount is set, compute shares = Amount/Price
		var qty decimal.Decimal
		if trade.Amount != nil {
			qty = trade.Amount.Div(price)
		} else if trade.Quantity != nil {
			qty = *trade.Quantity
		} else {
			qty = decimal.NewFromFloat(1)
			amt := qty.Mul(price)
			trade.Amount = &amt
		}
		trade.Price = &price
		trade.Quantity = &qty
		return trade, nil
	}

	switch trade.Action {
	case "BUY":
		order, err := Alpaca.PlaceOrder(a.PlaceOrderRequest{
			Side:          a.Side("buy"),
			Type:          a.OrderType("market"),
			Notional:      trade.Amount,
			Symbol:        trade.Symbol,
			ClientOrderID: trade.OrderID,
			TimeInForce:   a.TimeInForce("day"),
		})
		if err != nil {
			return nil, err
		}

		// Poll until the order reaches a terminal state (preferably filled)
		deadline := time.Now().Add(15 * time.Second)
		partialSeenAt := time.Time{}
		partialGrace := 3 * time.Second
		for {
			if order.Status == "filled" {
				// update the trade object with the market data
				trade.Price = order.FilledAvgPrice // price per share
				trade.Quantity = &order.FilledQty  // number of shares
				// ensure Amount reflects actual fill: price * quantity (decimal math)
				if trade.Price != nil && trade.Quantity != nil {
					amt := trade.Price.Mul(*trade.Quantity)
					trade.Amount = &amt
				}
				return trade, nil
			}

			// If the order is partially filled, allow a short grace window to complete
			// then accept the partial fill if there is a non-zero filled quantity
			if order.Status == "partially_filled" {
				if partialSeenAt.IsZero() {
					partialSeenAt = time.Now()
				} else if time.Since(partialSeenAt) >= partialGrace && order.FilledQty.GreaterThan(decimal.Zero) {
					trade.Price = order.FilledAvgPrice
					trade.Quantity = &order.FilledQty
					if trade.Price != nil && trade.Quantity != nil {
						amt := trade.Price.Mul(*trade.Quantity)
						trade.Amount = &amt
					}
					return trade, nil
				}
			}
			if order.Status == "canceled" || order.Status == "rejected" || order.Status == "expired" {
				return nil, fmt.Errorf("order %s ended with status %s", order.ID, order.Status)
			}
			if time.Now().After(deadline) {
				// On timeout, accept partial fills if any shares were executed
				if order.FilledQty.GreaterThan(decimal.Zero) {
					trade.Price = order.FilledAvgPrice
					trade.Quantity = &order.FilledQty
					if trade.Price != nil && trade.Quantity != nil {
						amt := trade.Price.Mul(*trade.Quantity)
						trade.Amount = &amt
					}
					return trade, nil
				}
				return nil, fmt.Errorf("timeout waiting for order %s to fill; last status %s", order.ID, order.Status)
			}
			time.Sleep(500 * time.Millisecond)
			// refresh order status
			refreshed, err := Alpaca.GetOrder(order.ID)
			if err != nil {
				// non-fatal: continue polling until deadline
				continue
			}
			order = refreshed
		}
	case "SELL":
		// TODO: Implement sell logic
		return nil, nil
	}
	return nil, nil
}
