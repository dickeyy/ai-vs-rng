package storage

import (
	"context"
	"fmt"

	"github.com/dickeyy/cis-320/services"
	"github.com/dickeyy/cis-320/types"
	"github.com/dickeyy/cis-320/utils"
	"github.com/rs/zerolog/log"
)

func SaveTrade(ctx context.Context, agentName string, trade *types.Trade) error {
	log.Debug().Str("order_id", trade.ID).Msg("Saving trade to database")

	// Check for nil pointers and provide default values
	var quantity, price, amount float64

	if trade.Quantity != nil {
		quantity, _ = trade.Quantity.Float64()
	} else {
		log.Debug().Str("order_id", trade.ID).Msg("Trade quantity is nil, using 0")
		quantity = 0
	}

	if trade.Price != nil {
		price, _ = trade.Price.Float64()
	} else {
		log.Debug().Str("order_id", trade.ID).Msg("Trade price is nil, using 0")
		price = 0
	}

	if trade.Amount != nil {
		amount, _ = trade.Amount.Float64()
	} else {
		log.Debug().Str("order_id", trade.ID).Msg("Trade amount is nil, using 0")
		amount = 0
	}

	// insert the trade into the database
	query := `
		INSERT INTO trades (id, agent_name, symbol, quantity, amount, price, action, timestamp, alpaca_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	id := trade.ID
	if id == "" {
		id = utils.GenerateOrderID()
		trade.ID = id
	}

	_, err := services.DB.ExecContext(
		ctx,
		query,
		id,
		agentName,
		trade.Symbol,
		quantity,
		amount,
		price,
		trade.Action,
		trade.Timestamp,
		trade.AlpacaID,
	)

	if err != nil {
		return fmt.Errorf("failed to insert trade into database: %w", err)
	}

	log.Debug().Str("order_id", trade.ID).Msg("Trade saved to database")
	return nil
}
