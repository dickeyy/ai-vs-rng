package agent

import (
	"context"
	"fmt"
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
	broker       types.Broker
	AlpacaClient *alpaca.Client
	tick         <-chan time.Time
}

var (
	lastTradeSymbol = ""
)

func NewRNGAgent(name string) *RNGStrategist {
	alpacaClient, account, holdings, err := services.InitializeAlpaca(os.Getenv("ALPACA_KEY_RNG"), os.Getenv("ALPACA_SECRET_RNG"))
	if err != nil {
		log.Fatal().Err(err).Msg("Error initializing Alpaca")
	}

	return &RNGStrategist{
		Name:    name,
		Symbols: utils.Symbols,
		AgentState: types.AgentState{
			Account:  *account,
			Holdings: holdings,
		},
		AlpacaClient: alpacaClient,
	}
}

// SetBroker sets the broker for the RNG Strategist.
func (a *RNGStrategist) SetBroker(broker types.Broker) {
	a.AgentState.Mu.Lock()
	defer a.AgentState.Mu.Unlock()
	a.broker = broker
}

// SetTickChannel sets the shared tick channel for the RNG Strategist.
func (a *RNGStrategist) SetTickChannel(tick <-chan time.Time) {
	a.AgentState.Mu.Lock()
	defer a.AgentState.Mu.Unlock()
	a.tick = tick
}

func (a *RNGStrategist) Run(ctx context.Context) error {
	// Use shared tick channel if provided, otherwise fall back to internal ticker
	var tickC <-chan time.Time
	if a.tick != nil {
		tickC = a.tick
	} else {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		tickC = ticker.C
	}

	for {
		select {
		case <-tickC:
			// make sure the market is open (only applies to non-dev mode)
			if !utils.IsTradingHours() {
				log.Debug().Str("agent", a.Name).Msg("Not trading hours, skipping tick")
				continue
			}

			// get a trade decision
			log.Info().Str("agent", a.Name).Msg("Making a random decision")
			a.updateAgentState()
			trade := a.makeDecision()

			// process trade
			if trade != nil {
				// check last trade symbol to avoid wash trading
				if trade.Symbol == lastTradeSymbol {
					log.Info().Str("agent", a.Name).Str("symbol", trade.Symbol).Msg("Skipping trade, last trade was the same symbol")
					continue
				}
				lastTradeSymbol = trade.Symbol

				// Submit the trade to the broker with a completion callback
				a.broker.SubmitTrade(ctx, trade, a.onComplete, a.AlpacaClient)

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

// makeDecision handles the agent's core algorithm
func (a *RNGStrategist) makeDecision() *types.Trade {
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
		trade := &types.Trade{
			ID:        utils.GenerateOrderID(),
			AlpacaID:  "", // to be set by the Alpaca service after placement
			Symbol:    symbol,
			Amount:    &tradeAmount,
			Action:    "BUY",
			Timestamp: time.Now(),
			AgentName: a.Name,
		}
		err := a.validateTrade(trade)
		if err != nil {
			log.Error().Err(err).Str("agent", a.Name).Msg("Error validating trade")
			return nil
		}
		return trade
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
		trade := &types.Trade{
			ID:        utils.GenerateOrderID(),
			AlpacaID:  "", // to be set by the Alpaca service after placement
			Symbol:    holding.Symbol,
			Quantity:  &tradeQuantity,
			Action:    "SELL",
			Timestamp: time.Now(),
			AgentName: a.Name,
		}
		err := a.validateTrade(trade)
		if err != nil {
			log.Error().Err(err).Str("agent", a.Name).Msg("Error validating trade")
			return nil
		}
		return trade
	} else {
		// Hold
		return nil
	}
}

func (a *RNGStrategist) onComplete(trade *types.Trade, processed *types.Trade, err error) {
	if err != nil {
		log.Error().Err(err).Str("agent", a.Name).Str("order_id", trade.ID).Msg("Trade failed or was rejected")
		return
	}

	if processed == nil {
		log.Error().Str("agent", a.Name).Str("order_id", trade.ID).Msg("Broker completed with nil trade")
		return
	}

	// wait for 5 seconds to make sure Alpaca fills the order
	time.Sleep(5 * time.Second)

	// perform state updates only after broker finished processing
	a.AgentState.Mu.Lock()
	defer a.AgentState.Mu.Unlock()
	a.updateAgentState()

	log.Info().Str("agent", a.Name).Str("order_id", processed.ID).Msg("State updated and saved for processed trade")
}

// updateAgentState updates the agent state with the latest account and holdings from Alpaca
func (a *RNGStrategist) updateAgentState() {
	// fetch the latest account and holdings from alpaca
	account, err := services.GetAccount(a.AlpacaClient)
	if err != nil {
		log.Error().Err(err).Msg("Error getting account")
	}
	holdings, err := services.GetHoldings(a.AlpacaClient)
	if err != nil {
		log.Error().Err(err).Msg("Error getting holdings")
	}

	// update the agent state
	a.AgentState.Account = *account
	a.AgentState.Holdings = holdings
}

func (a *RNGStrategist) getHolding(symbol string) (alpaca.Position, error) {
	for _, holding := range a.AgentState.Holdings {
		if holding.Symbol == symbol {
			return holding, nil
		}
	}
	return alpaca.Position{}, fmt.Errorf("holding not found")
}

// validateTradeDecision validates the trade decision
func (a *RNGStrategist) validateTrade(trade *types.Trade) error {
	if trade.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}

	switch trade.Action {
	case "BUY":
		if trade.Amount == nil || trade.Amount.IsZero() {
			return fmt.Errorf("amount is required")
		}
		if trade.Amount.GreaterThan(a.AgentState.Account.BuyingPower) {
			return fmt.Errorf("amount is greater than buying power")
		}
	case "SELL":
		if trade.Quantity == nil || trade.Quantity.IsZero() {
			return fmt.Errorf("quantity is required")
		}
		// get the holding for the symbol
		holding, err := a.getHolding(trade.Symbol)
		if err != nil {
			return fmt.Errorf("holding not found")
		}
		if trade.Quantity.GreaterThan(holding.QtyAvailable) {
			return fmt.Errorf("quantity is greater than available quantity")
		}
	}

	return nil
}
