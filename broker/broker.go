package broker

import (
	"context"
	"sync"
	"time"

	"github.com/dickeyy/cis-320/services"
	"github.com/dickeyy/cis-320/storage"
	"github.com/dickeyy/cis-320/types"
	"github.com/dickeyy/cis-320/utils"
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

// workItem is a unit of work for the broker queue containing the trade and an optional completion callback.
type workItem struct {
	trade      *types.Trade
	onComplete func(*types.Trade, error)
}

// SubmitTrade adds a trade to the broker's queue for processing.
func (b *Broker) SubmitTrade(ctx context.Context, trade *types.Trade, onComplete func(*types.Trade, error)) {
	b.tradeQueue.Enqueue(&workItem{trade: trade, onComplete: onComplete})
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
					wi := b.tradeQueue.Dequeue()
					if wi != nil {
						trade := wi.trade
						if trade.ID == "" {
							trade.ID = utils.GenerateOrderID()
						}
						log.Info().Str("order_id", trade.ID).Msg("Broker processing trade")
						// submit the trade to the alpaca api
						updatedTrade, err := services.PlaceOrder(trade)
						if err != nil {
							log.Error().Err(err).Str("order_id", trade.ID).Msg("Error placing order")
							if wi.onComplete != nil {
								wi.onComplete(nil, err)
							}
						} else {
							trade = updatedTrade
							log.Info().Str("order_id", trade.ID).Msg("Order placed successfully")

							// save the trade to the database (works with both successful and failed orders)
							err = storage.SaveTrade(ctx, trade.AgentName, trade)
							if err != nil {
								log.Error().Err(err).Str("order_id", trade.ID).Msg("Error saving trade to DB")
							}
							if wi.onComplete != nil {
								wi.onComplete(trade, nil)
							}
						}
						log.Info().Str("order_id", trade.ID).Msg("Trade processed")
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
