package broker

import (
	"context"
	"sync"
	"time"

	"github.com/dickeyy/cis-320/services"
	"github.com/dickeyy/cis-320/storage"
	"github.com/dickeyy/cis-320/types"
	"github.com/rs/zerolog/log"
)

// Broker handles the execution of trades submitted by agents.
type Broker struct {
	mu         sync.Mutex
	tradeQueue *TradeQueue
}

// NewBroker creates and returns a new Broker.
func NewBroker() *Broker {
	return &Broker{
		tradeQueue: NewTradeQueue(),
	}
}

// SubmitTrade adds a trade to the broker's queue for processing.
func (b *Broker) SubmitTrade(ctx context.Context, trade *types.Trade) {
	b.tradeQueue.Enqueue(trade)
}

// ProcessTrades starts a goroutine to continuously process trades from the queue.
func (b *Broker) ProcessTrades(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Info().Msg("Broker shutting down trade processing.")
				return
			default:
				if !b.tradeQueue.IsEmpty() {
					trade := b.tradeQueue.Dequeue()
					if trade != nil {
						log.Info().Str("order_id", trade.OrderID).Msg("Broker processing trade")
						// submit the trade to the alpaca api
						updatedTrade, err := services.PlaceOrder(trade)
						if err != nil {
							log.Error().Msgf("Error placing order: %v", err)
						} else {
							trade = updatedTrade
							log.Info().Str("order_id", trade.OrderID).Msg("Order placed successfully")

							// save the trade to the database (works with both successful and failed orders)
							err = storage.SaveTrade(ctx, trade.AgentName, trade)
							if err != nil {
								log.Error().Msgf("Error saving trade to DB: %v", err)
							}
						}
						log.Info().Str("order_id", trade.OrderID).Msg("Trade processed")
					} else {
						log.Debug().Msg("Trade queue was empty after check, but Dequeue returned nil.")
					}
				} else {
					// small delay to avoid busy-waiting
					time.Sleep(100 * time.Millisecond)
				}
			}
		}
	}()
}
