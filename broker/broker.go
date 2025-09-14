package broker

import (
	"context"
	"sync"
	"time"

	"github.com/dickeyy/cis-320/services"
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
	apiKey     string
	apiSecret  string
}

// SubmitTrade adds a trade to the broker's queue for processing.
func (b *Broker) SubmitTrade(ctx context.Context, trade *types.Trade, onComplete func(*types.Trade, error), apiKey, apiSecret string) {
	b.tradeQueue.Enqueue(&workItem{trade: trade, onComplete: onComplete, apiKey: apiKey, apiSecret: apiSecret})
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

						processedTrade, err := services.PlaceOrder(trade, wi.apiKey, wi.apiSecret)
						if err != nil {
							log.Error().Err(err).Str("order_id", trade.ID).Msg("Error placing order")
							if wi.onComplete != nil {
								wi.onComplete(nil, err)
							}
						} else {
							if processedTrade != nil {
								log.Info().Str("order_id", processedTrade.ID).Str("alpaca_id", processedTrade.AlpacaID).Msg("Order placed successfully")
							} else {
								log.Info().Str("order_id", trade.ID).Msg("Order placed successfully")
							}
							if wi.onComplete != nil {
								// prefer returning processed trade if available
								if processedTrade != nil {
									wi.onComplete(processedTrade, nil)
								} else {
									wi.onComplete(trade, nil)
								}
							}
						}
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
