package utils

import (
	"os"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/rs/zerolog/log"
)

func IsTradingHours() bool {
	if DevMode {
		return true
	}

	open, err := alpaca.NewClient(alpaca.ClientOpts{
		// use RNG api creds (doesnt really matter we just need some valid creds)
		APIKey:    os.Getenv("ALPACA_KEY"),
		APISecret: os.Getenv("ALPACA_SECRET"),
		BaseURL:   "https://paper-api.alpaca.markets",
	}).GetClock()

	if err != nil {
		log.Error().Err(err).Msg("Error getting clock")
		return false
	}

	return open.IsOpen
}
