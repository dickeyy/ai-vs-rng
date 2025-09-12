package services

import (
	"fmt"
	"os"

	a "github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/dickeyy/cis-320/types"
)

var (
	Alpaca        *a.Client  = nil
	AlpacaAccount *a.Account = nil
)

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

		if order.Status != "accepted" && order.Status != "filled" {
			return nil, fmt.Errorf("order status is %s", order.Status)
		}

		// update the trade object with the market data
		trade.Price = order.FilledAvgPrice // price per share
		trade.OrderID = order.ID           // order id
		trade.Quantity = &order.FilledQty  // number of shares

		return trade, nil
	case "SELL":
		// TODO: Implement sell logic
		return nil, nil
	}
	return nil, nil
}
