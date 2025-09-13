package agent

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/dickeyy/cis-320/services"
	"github.com/dickeyy/cis-320/types"
	"github.com/dickeyy/cis-320/utils"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

// RNGStrategist is an agent that makes random buy/sell/hold decisions.
type RNGStrategist struct {
	Name    string
	Symbols []string
	// Embed AgentState to manage common agent properties
	types.AgentState
	broker types.Broker
}

func NewRNGAgent(name string, symbols []string) *RNGStrategist {
	return &RNGStrategist{
		Name:    name,
		Symbols: symbols,
		AgentState: types.AgentState{
			Name:     name,
			Holdings: []types.Position{},
			Stats: types.AgentStats{
				InitialBalance: decimal.NewFromFloat(10000), // Starting capital
				CurrentBalance: decimal.NewFromFloat(10000),
				TotalTrades:    0,
				WinningTrades:  0,
				LosingTrades:   0,
			},
		},
	}
}

// SetBroker sets the broker for the RNG Strategist.
func (a *RNGStrategist) SetBroker(broker types.Broker) {
	a.AgentState.Mu.Lock()
	defer a.AgentState.Mu.Unlock()
	a.broker = broker
}

func (a *RNGStrategist) Run(ctx context.Context) error {
	// attempt to load state
	err := a.LoadState(ctx)
	if err != nil {
		if err.Error() == "failed to load agent state: no data found for key RNG_Agent" {
			log.Info().Str("agent", a.Name).Msg("No state found, starting fresh")
		} else {
			return err
		}
	}
	log.Info().Str("agent", a.Name).Msg("Loaded state")

	// Simulate a ticker within the agent
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Info().Str("agent", a.Name).Msg("Making a random decision")
			trade := a.makeRandomDecision()
			if trade != nil {
				// Submit the trade to the broker with a completion callback
				a.broker.SubmitTrade(ctx, trade, func(processed *types.Trade, err error) {
					if err != nil {
						log.Error().Err(err).Str("agent", a.Name).Str("order_id", trade.ID).Msg("Trade failed or was rejected")
						return
					}
					if processed == nil {
						log.Error().Str("agent", a.Name).Str("order_id", trade.ID).Msg("Broker completed with nil trade")
						return
					}
					// perform state updates only after broker finished processing
					a.AgentState.Mu.Lock()
					defer a.AgentState.Mu.Unlock()
					a.updateBasicStats(processed)
					a.updateHoldings(processed)
					_ = a.SaveState(ctx)
					log.Info().Str("agent", a.Name).Str("order_id", processed.ID).Msg("State updated and saved for processed trade")
				})
				log.Info().Str("agent", a.Name).Str("action", trade.Action).Str("order_id", trade.ID).Msg("Submitted order to broker")
			} else {
				log.Info().Str("agent", a.Name).Msg("No trade made")
			}
		case <-ctx.Done():
			log.Info().Str("agent", a.Name).Msg("Shutting down RNG Agent")
			return nil
		}
	}
}

func (a *RNGStrategist) Stop(ctx context.Context) error {
	err := a.SaveState(ctx)
	if err != nil {
		return fmt.Errorf("failed to save agent state: %w", err)
	}
	log.Info().Str("agent", a.Name).Msg("Saved state on stop")
	return nil
}

// GetName returns the name of the RNG Strategist.
func (a *RNGStrategist) GetName() string {
	return a.AgentState.Name
}

// GetHoldings returns the agent current holdings
func (a *RNGStrategist) GetHoldings(ctx context.Context) ([]types.Position, error) {
	return a.AgentState.Holdings, nil
}

// GetCashBalance returns the agent current cash balance
func (a *RNGStrategist) GetCashBalance(ctx context.Context) (decimal.Decimal, error) {
	return a.AgentState.Stats.CurrentBalance, nil
}

// GetCurrentPortfolioValue returns the agent current portfolio value (cash + holdings value)
func (a *RNGStrategist) GetCurrentPortfolioValue(ctx context.Context) (decimal.Decimal, error) {
	holdingsValue := decimal.NewFromFloat(0)
	for _, holding := range a.AgentState.Holdings {
		holdingsValue = holdingsValue.Add(holding.MarketValue)
	}
	return a.AgentState.Stats.CurrentBalance.Add(holdingsValue), nil
}

// GetStats returns the agent current stats
func (a *RNGStrategist) GetStats(ctx context.Context) (types.AgentStats, error) {
	return a.AgentState.Stats, nil
}

// SaveState saves the agent current state to redis
func (a *RNGStrategist) SaveState(ctx context.Context) error {
	log.Debug().Str("agent", a.Name).Msg("Saving state")
	err := services.SaveAgentState(ctx, a.Name, &a.AgentState)
	if err != nil {
		return fmt.Errorf("failed to save agent state: %w", err)
	}
	return nil
}

// LoadState loads the agent last saved state from redis
func (a *RNGStrategist) LoadState(ctx context.Context) error {
	log.Debug().Str("agent", a.Name).Msg("Loading state")
	err := services.LoadAgentState(ctx, a.Name, &a.AgentState)
	if err != nil {
		return fmt.Errorf("failed to load agent state: %w", err)
	}
	return nil
}

// makeRandomDecision handles the agent's core algorithm
func (a *RNGStrategist) makeRandomDecision() *types.Trade {
	// get a number between 1-100
	r := utils.RNG(1, 100)
	log.Debug().Str("agent", a.Name).Int("random_value", r).Msg("Random number generated")

	// each option will have a 33% chance
	// Buy, Sell, or Hold
	if r <= 33 {
		// Buy
		// choose a random symbol
		symbol := utils.RandomString(a.Symbols)
		// based on the agents capital, choose a random value < current capital
		currBalance, _ := a.AgentState.Stats.CurrentBalance.Float64()
		spend := math.Floor(utils.RandomFloat(0, currBalance)*100+0.5) / 100
		log.Debug().Str("agent", a.Name).Str("symbol", symbol).Float64("amount", spend).Msg("Buying")

		// make the base trade object, this will be updated later with real market data by the broker
		var tradeAmount = decimal.NewFromFloat(spend)
		return &types.Trade{
			ID:        utils.GenerateOrderID(),
			AlpacaID:  "", // to be set by the Alpaca service after placement
			Symbol:    symbol,
			Amount:    &tradeAmount,
			Action:    "BUY",
			Timestamp: time.Now(),
			AgentName: a.Name,
		}
	} else if r <= 66 {
		// Sell
		log.Debug().Str("agent", a.Name).Msg("Selling")
	} else {
		// Hold
		return nil
	}

	return nil
}

func (a *RNGStrategist) updateBasicStats(trade *types.Trade) {
	if trade.Amount != nil {
		if trade.Action == "BUY" {
			a.AgentState.Stats.CurrentBalance = a.AgentState.Stats.CurrentBalance.Sub(*trade.Amount)
		} else {
			a.AgentState.Stats.CurrentBalance = a.AgentState.Stats.CurrentBalance.Add(*trade.Amount)
		}
	}
	a.AgentState.Stats.TotalTrades++
}

func (a *RNGStrategist) updateHoldings(trade *types.Trade) {
	if trade.Quantity == nil || trade.Price == nil || trade.Amount == nil {
		// nothing to do without complete execution details
		return
	}
	if len(a.AgentState.Holdings) == 0 {
		a.AgentState.Holdings = append(a.AgentState.Holdings, types.Position{
			Symbol:      trade.Symbol,
			Quantity:    *trade.Quantity,
			CPS:         *trade.Price,
			MarketValue: *trade.Amount,
		})
	} else {
		for i := range a.AgentState.Holdings {
			holding := &a.AgentState.Holdings[i]
			if holding.Symbol == trade.Symbol {
				if trade.Action == "BUY" {
					holding.Quantity = holding.Quantity.Add(*trade.Quantity)
					holding.CPS = *trade.Price // trade.Price should be the current price per share (prov. by alpaca)
					holding.MarketValue = holding.CPS.Mul(holding.Quantity)
				} else {
					holding.Quantity = holding.Quantity.Sub(*trade.Quantity)
					holding.CPS = *trade.Price // trade.Price should be the current price per share (prov. by alpaca)
					holding.MarketValue = holding.CPS.Mul(holding.Quantity)
				}
				return
			}
		}
		// if we didn't find an existing holding for the symbol, add new
		a.AgentState.Holdings = append(a.AgentState.Holdings, types.Position{
			Symbol:      trade.Symbol,
			Quantity:    *trade.Quantity,
			CPS:         *trade.Price,
			MarketValue: *trade.Amount,
		})
	}
}
