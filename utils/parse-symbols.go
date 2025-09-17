package utils

import (
	"fmt"
	"os"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
)

var (
	Symbols []string
)

func ParseSymbols() error {
	d, err := alpaca.NewClient(alpaca.ClientOpts{
		APIKey:    os.Getenv("ALPACA_KEY_RNG"),
		APISecret: os.Getenv("ALPACA_SECRET_RNG"),
		BaseURL:   "https://paper-api.alpaca.markets",
	}).GetAssets(alpaca.GetAssetsRequest{
		Status:     "active",
		AssetClass: "us_equity",
	})
	if err != nil {
		return err
	}

	symbols := make([]string, 0, len(d))
	for _, asset := range d {
		if asset.Fractionable && asset.Tradable && asset.Status == alpaca.AssetActive {
			symbols = append(symbols, asset.Symbol)
		}
	}

	if len(symbols) == 0 {
		return fmt.Errorf("no symbols found")
	}

	Symbols = symbols
	return nil
}
