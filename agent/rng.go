package agent

import (
	"context"
	"math"
	"os"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
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

var (
	apiKey          = ""
	apiSecret       = ""
	lastTradeSymbol = ""
)

func NewRNGAgent(name string, symbols []string) *RNGStrategist {
	apiKey = os.Getenv("ALPACA_KEY")
	apiSecret = os.Getenv("ALPACA_SECRET")

	if apiKey == "" || apiSecret == "" {
		log.Fatal().Msg("ALPACA_KEY/ALPACA_SECRET not set")
	}
	log.Info().Msg("Alpaca credentials loaded from environment")

	account, err := services.GetAccount(apiKey, apiSecret)
	if err != nil {
		log.Fatal().Err(err).Msg("Error getting account")
	}
	holdings, err := services.GetHoldings(apiKey, apiSecret)
	if err != nil {
		log.Fatal().Err(err).Msg("Error getting holdings")
	}

	return &RNGStrategist{
		Name:    name,
		Symbols: symbols,
		AgentState: types.AgentState{
			Account:  *account,
			Holdings: holdings,
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
	// Simulate a ticker within the agent
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// make sure the market is open (only applies to non-dev mode)
			if !utils.IsTradingHours() {
				log.Debug().Str("agent", a.Name).Msg("Not trading hours, skipping")
				continue
			}

			// process trade
			log.Info().Str("agent", a.Name).Msg("Making a random decision")
			trade := a.makeRandomDecision()
			if trade != nil {
				// check last trade symbolt to avoid wash trading
				if trade.Symbol == lastTradeSymbol {
					log.Info().Str("agent", a.Name).Str("symbol", trade.Symbol).Msg("Skipping trade, last trade was the same symbol")
					continue
				}
				lastTradeSymbol = trade.Symbol

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
					a.updateAgentState()

					log.Info().Str("agent", a.Name).Str("order_id", processed.ID).Msg("State updated and saved for processed trade")
				}, apiKey, apiSecret)

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
	return nil
}

// GetName returns the name of the RNG Strategist.
func (a *RNGStrategist) GetName() string {
	return a.Name
}

// GetHoldings returns the agent current holdings
func (a *RNGStrategist) GetHoldings(ctx context.Context) ([]alpaca.Position, error) {
	return a.AgentState.Holdings, nil
}

// GetBuyingPower returns the agent current buying power
func (a *RNGStrategist) GetBuyingPower(ctx context.Context) (decimal.Decimal, error) {
	return a.AgentState.Account.BuyingPower, nil
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
		// based on the agents capital, choose a random value <= current capital
		currBalance, _ := a.AgentState.Account.BuyingPower.Float64()
		spend := math.Floor(utils.RandomFloat(0, currBalance)*100+0.5) / 100
		// clamp the spend to the current balance
		spend = math.Min(spend, currBalance)
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
		// choose a random holding
		holding := utils.RandomItem(a.AgentState.Holdings)
		// based on the holding's quantity, choose a random value <= quantity
		qty, _ := holding.QtyAvailable.Float64()
		sell := math.Floor(utils.RandomFloat(0, qty)*100+0.5) / 100
		sell = math.Min(sell, qty)
		log.Debug().Str("agent", a.Name).Str("holding", holding.Symbol).Float64("amount", sell).Msg("Selling")

		// make the base trade object, this will be updated later with real market data by the broker
		var tradeQuantity = decimal.NewFromFloat(sell)
		return &types.Trade{
			ID:        utils.GenerateOrderID(),
			AlpacaID:  "", // to be set by the Alpaca service after placement
			Symbol:    holding.Symbol,
			Quantity:  &tradeQuantity,
			Action:    "SELL",
			Timestamp: time.Now(),
			AgentName: a.Name,
		}
	} else {
		// Hold
		return nil
	}
}

func (a *RNGStrategist) updateAgentState() {
	// fetch the latest account and holdings from alpaca
	account, err := services.GetAccount(apiKey, apiSecret)
	if err != nil {
		log.Error().Err(err).Msg("Error getting account")
	}
	holdings, err := services.GetHoldings(apiKey, apiSecret)
	if err != nil {
		log.Error().Err(err).Msg("Error getting holdings")
	}

	// update the agent state
	a.AgentState.Account = *account
	a.AgentState.Holdings = holdings
}
