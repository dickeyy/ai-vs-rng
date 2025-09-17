package services

import (
	"fmt"
	"os"

	a "github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/dickeyy/cis-320/types"
)

func InitializeAlpaca(apiKey, apiSecret string) (*a.Client, *a.Account, []a.Position, error) {
	baseURL := os.Getenv("ALPACA_API")

	if baseURL == "" {
		baseURL = "https://paper-api.alpaca.markets"
	}
	if apiKey == "" || apiSecret == "" {
		return nil, nil, nil, fmt.Errorf("ALPACA_KEY and ALPACA_SECRET must be set for live mode")
	}

	client := a.NewClient(a.ClientOpts{
		APIKey:    apiKey,
		APISecret: apiSecret,
		BaseURL:   baseURL,
	})

	account, err := GetAccount(client)
	if err != nil {
		return nil, nil, nil, err
	}

	holdings, err := GetHoldings(client)
	if err != nil {
		return nil, nil, nil, err
	}

	return client, account, holdings, nil
}

func PlaceOrder(trade *types.Trade, client *a.Client) (*types.Trade, error) {
	switch trade.Action {
	case "BUY":
		order, err := client.PlaceOrder(a.PlaceOrderRequest{
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
	case "SELL":
		order, err := client.PlaceOrder(a.PlaceOrderRequest{
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
	}

	return trade, nil
}

func GetHoldings(client *a.Client) ([]a.Position, error) {
	holdings, err := client.GetPositions()
	if err != nil {
		return nil, err
	}

	return holdings, nil
}

func GetAccount(client *a.Client) (*a.Account, error) {
	account, err := client.GetAccount()
	if err != nil {
		return nil, err
	}

	return account, nil
}
