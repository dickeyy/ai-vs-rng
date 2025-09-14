package services

import (
	"fmt"
	"os"

	a "github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	md "github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	"github.com/dickeyy/cis-320/types"
)

var (
	Alpaca        *a.Client  = nil
	AlpacaAccount *a.Account = nil
	DataAPI       *md.Client = nil
)

func InitializeAlpaca(apiKey, apiSecret string) error {
	baseURL := os.Getenv("ALPACA_API")

	if baseURL == "" {
		baseURL = "https://paper-api.alpaca.markets"
	}
	if apiKey == "" || apiSecret == "" {
		return fmt.Errorf("ALPACA_KEY and ALPACA_SECRET must be set for live mode")
	}

	Alpaca = a.NewClient(a.ClientOpts{
		APIKey:    apiKey,
		APISecret: apiSecret,
		BaseURL:   baseURL,
	})
	acct, err := Alpaca.GetAccount()
	if err != nil {
		return err
	}
	AlpacaAccount = acct

	return nil
}

func ShutdownAlpaca() {
	Alpaca = nil
	AlpacaAccount = nil
}

func PlaceOrder(trade *types.Trade, apiKey, apiSecret string) (*types.Trade, error) {
	// initalize alpaca
	err := InitializeAlpaca(apiKey, apiSecret)
	if err != nil {
		return nil, err
	}

	switch trade.Action {
	case "BUY":
		order, err := Alpaca.PlaceOrder(a.PlaceOrderRequest{
			Side:        a.Side("buy"),
			Type:        a.OrderType("market"),
			Notional:    trade.Amount,
			Symbol:      trade.Symbol,
			TimeInForce: a.TimeInForce("day"),
		})
		if err != nil {
			return nil, err
		}

		// Persist the Alpaca order id onto the trade for downstream usage
		trade.AlpacaID = order.ID

		ShutdownAlpaca()
	case "SELL":
		order, err := Alpaca.PlaceOrder(a.PlaceOrderRequest{
			Side:        a.Side("sell"),
			Type:        a.OrderType("market"),
			Qty:         trade.Quantity,
			Symbol:      trade.Symbol,
			TimeInForce: a.TimeInForce("day"),
		})
		if err != nil {
			return nil, err
		}

		// Persist the Alpaca order id onto the trade for downstream usage
		trade.AlpacaID = order.ID

		ShutdownAlpaca()
	}
	return trade, nil
}

func GetHoldings(apiKey, apiSecret string) ([]a.Position, error) {
	err := InitializeAlpaca(apiKey, apiSecret)
	if err != nil {
		return nil, err
	}

	holdings, err := Alpaca.GetPositions()
	if err != nil {
		return nil, err
	}

	ShutdownAlpaca()

	return holdings, nil
}

func GetAccount(apiKey, apiSecret string) (*a.Account, error) {
	err := InitializeAlpaca(apiKey, apiSecret)
	if err != nil {
		return nil, err
	}

	account, err := Alpaca.GetAccount()
	if err != nil {
		return nil, err
	}

	ShutdownAlpaca()

	return account, nil
}
