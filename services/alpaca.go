package services

import (
	"os"

	a "github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
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
