package agent

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/dickeyy/cis-320/services"
	"github.com/dickeyy/cis-320/types"
	"github.com/dickeyy/cis-320/utils"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

// LLMStrategist is an agent that makes buy/sell/hold decisions based on a LLM.
type LLMStrategist struct {
	Name    string
	Symbols []string
	// Embed AgentState to manage common agent properties
	types.AgentState
	broker       types.Broker
	AlpacaClient *alpaca.Client
	tick         <-chan time.Time
	LastError    error
}

func NewLLMAgent(name string) *LLMStrategist {
	alpacaClient, account, holdings, err := services.InitializeAlpaca(os.Getenv("ALPACA_KEY_LLM"), os.Getenv("ALPACA_SECRET_LLM"))
	if err != nil {
		log.Fatal().Err(err).Msg("Error initializing Alpaca")
	}

	return &LLMStrategist{
		Name:    name,
		Symbols: utils.Symbols,
		AgentState: types.AgentState{
			Account:  *account,
			Holdings: holdings,
		},
		AlpacaClient: alpacaClient,
	}
}

// SetBroker sets the broker for the LLM agent
func (a *LLMStrategist) SetBroker(broker types.Broker) {
	a.AgentState.Mu.Lock()
	defer a.AgentState.Mu.Unlock()
	a.broker = broker
}

// SetTickChannel sets the shared tick channel for the LLM agent
func (a *LLMStrategist) SetTickChannel(tick <-chan time.Time) {
	a.AgentState.Mu.Lock()
	defer a.AgentState.Mu.Unlock()
	a.tick = tick
}

// Run starts the LLM agent's primary trading loop
func (a *LLMStrategist) Run(ctx context.Context) error {
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
			// make sure the market is open
			if !utils.IsTradingHours() {
				log.Debug().Str("agent", a.Name).Msg("Not trading hours, skipping tick")
				continue
			}

			// get a trade decision
			log.Info().Str("agent", a.Name).Msg("Making a decision")
			a.updateAgentState()
			trade := a.makeDecision(ctx)

			// process trade
			if trade != nil {
				// submit the trade to the broker with a completion callback
				a.broker.SubmitTrade(ctx, trade, a.onComplete, a.AlpacaClient)

				log.Info().Str("agent", a.Name).Str("action", trade.Action).Str("order_id", trade.ID).Msg("Submitted order to broker")
			} else {
				log.Info().Str("agent", a.Name).Msg("No trade made")
				holdTrade := types.Trade{
					ID:        utils.GenerateOrderID(),
					AlpacaID:  "",
					Symbol:    "",
					Amount:    nil,
					Quantity:  nil,
					Action:    "HOLD",
					Timestamp: time.Now(),
					AgentName: a.Name,
				}
				a.onComplete(nil, &holdTrade, nil)
			}
		case <-ctx.Done():
			log.Info().Str("agent", a.Name).Msg("Shutting down LLM Agent")
			return nil
		}

	}
}

func (a *LLMStrategist) Stop(ctx context.Context) error {
	return nil
}

// GetName returns the name of the RNG Strategist.
func (a *LLMStrategist) GetName() string {
	return a.Name
}

// GetHoldings returns the agent current holdings
func (a *LLMStrategist) GetHoldings(ctx context.Context) ([]alpaca.Position, error) {
	return a.AgentState.Holdings, nil
}

// GetBuyingPower returns the agent current buying power
func (a *LLMStrategist) GetBuyingPower(ctx context.Context) (decimal.Decimal, error) {
	return a.AgentState.Account.BuyingPower, nil
}

// makeDecision handles the agent's core algorithm
func (a *LLMStrategist) makeDecision(ctx context.Context) *types.Trade {
	// snapshot state under lock
	a.AgentState.Mu.Lock()
	account := a.AgentState.Account
	holdings := make([]alpaca.Position, len(a.AgentState.Holdings))
	copy(holdings, a.AgentState.Holdings)
	lastError := a.LastError
	a.AgentState.Mu.Unlock()

	// create temp state for AI call
	tempState := &types.AgentState{
		Account:  account,
		Holdings: holdings,
	}

	// get a trade decision from the ai
	tradeDecision, err := services.GetAITradeDecision(ctx, tempState, lastError)
	if err != nil {
		log.Error().Err(err).Str("agent", a.Name).Msg("Error getting AI trade decision")
		return nil
	}

	// validate the trade
	err = a.validateTradeDecision(tradeDecision)
	if err != nil && tradeDecision.Action != "NONE" {
		log.Error().Err(err).Str("agent", a.Name).Msg("Error validating trade decision")
		return nil
	}

	tradeID := utils.GenerateOrderID()
	err = services.SaveAIReasoning(a.Name, tradeDecision.Reasoning, tradeID, ctx)
	if err != nil {
		log.Error().Err(err).Str("agent", a.Name).Msg("Error saving AI reasoning")
		return nil
	}

	switch tradeDecision.Action {
	case "BUY":
		return &types.Trade{
			ID:        tradeID,
			AlpacaID:  "",
			Symbol:    tradeDecision.Symbol,
			Amount:    tradeDecision.Amount,
			Action:    tradeDecision.Action,
			Timestamp: time.Now(),
			AgentName: a.Name,
		}
	case "SELL":
		return &types.Trade{
			ID:        tradeID,
			AlpacaID:  "",
			Symbol:    tradeDecision.Symbol,
			Quantity:  tradeDecision.Quantity,
			Action:    tradeDecision.Action,
			Timestamp: time.Now(),
			AgentName: a.Name,
		}
	default:
		return nil
	}
}

func (a *LLMStrategist) onComplete(trade *types.Trade, processed *types.Trade, err error) {
	if err != nil {
		// Save the error for future decision making
		a.AgentState.Mu.Lock()
		a.LastError = err
		a.AgentState.Mu.Unlock()

		if trade != nil {
			log.Error().Err(err).Str("agent", a.Name).Str("order_id", trade.ID).Msg("Trade failed or was rejected")
		} else {
			log.Error().Err(err).Str("agent", a.Name).Msg("Trade failed or was rejected")
		}
		return
	}

	if processed == nil {
		if trade != nil {
			log.Error().Str("agent", a.Name).Str("order_id", trade.ID).Msg("Broker completed with nil trade")
		} else {
			log.Error().Str("agent", a.Name).Msg("Broker completed with nil trade")
		}
		return
	}

	// wait for 5 seconds to make sure Alpaca fills the order
	time.Sleep(5 * time.Second)

	// perform state updates only after broker finished processing
	a.AgentState.Mu.Lock()
	defer a.AgentState.Mu.Unlock()
	a.updateAgentState()

	// Clear any previous error since this trade succeeded
	a.LastError = nil

	err = services.SaveTrade(processed, context.Background())
	if err != nil {
		log.Error().Err(err).Str("agent", a.Name).Str("order_id", processed.ID).Msg("Error saving trade")
	}

	log.Info().Str("agent", a.Name).Str("order_id", processed.ID).Msg("State updated and saved for processed trade")
}

// updateAgentState updates the agent state with the latest account and holdings from Alpaca
func (a *LLMStrategist) updateAgentState() {
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

func (a *LLMStrategist) getHolding(symbol string) (alpaca.Position, error) {
	for _, holding := range a.AgentState.Holdings {
		if holding.Symbol == symbol {
			return holding, nil
		}
	}
	return alpaca.Position{}, fmt.Errorf("holding not found")
}

// validateTradeDecision validates the trade decision
func (a *LLMStrategist) validateTradeDecision(tradeDecision *types.TradeDecision) error {
	if tradeDecision.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}

	switch tradeDecision.Action {
	case "BUY":
		if tradeDecision.Amount == nil || tradeDecision.Amount.IsZero() {
			return fmt.Errorf("amount is required")
		}
		if tradeDecision.Amount.GreaterThan(a.AgentState.Account.BuyingPower) {
			return fmt.Errorf("amount is greater than buying power")
		}
	case "SELL":
		if tradeDecision.Quantity == nil || tradeDecision.Quantity.IsZero() {
			return fmt.Errorf("quantity is required")
		}
		// get the holding for the symbol
		holding, err := a.getHolding(tradeDecision.Symbol)
		if err != nil {
			return fmt.Errorf("holding not found")
		}
		if tradeDecision.Quantity.GreaterThan(holding.QtyAvailable) {
			return fmt.Errorf("quantity is greater than available quantity")
		}
	}

	return nil
}
