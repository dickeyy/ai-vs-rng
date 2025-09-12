package types

import (
	"context"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

// Trade represents a single trade made by an agent
type Trade struct {
	Symbol    string           `json:"symbol"`     // stock ticker, e.g. "APPL"
	Quantity  *decimal.Decimal `json:"quantity"`   // number of shares
	Amount    *decimal.Decimal `json:"amount"`     // amount of the trade
	Price     *decimal.Decimal `json:"price"`      // price per share
	Action    string           `json:"action"`     // "BUY" or "SELL"
	Timestamp time.Time        `json:"timestamp"`  // time of the trade
	OrderID   string           `json:"order_id"`   // unique identifier for the trade
	AgentName string           `json:"agent_name"` // name of the agent that made the trade
}

// Position represents a current position in a stock
type Position struct {
	Symbol      string          `json:"symbol"`
	Quantity    decimal.Decimal `json:"quantity"`
	CPS         decimal.Decimal `json:"cps"`          // current price per share
	MarketValue decimal.Decimal `json:"market_value"` // current market value of the position (CPS * quantity)
}

// AgentStats holds key performance indicators for an agent
type AgentStats struct {
	CurrentBalance decimal.Decimal `json:"current_balance"`
	InitialBalance decimal.Decimal `json:"initial_balance"`
	ProfitLoss     float64         `json:"profit_loss"` // Overall P/L
	ROI            decimal.Decimal `json:"roi"`         // Return on Investment
	TotalTrades    int             `json:"total_trades"`
	WinningTrades  int             `json:"winning_trades"`
	LosingTrades   int             `json:"losing_trades"`
}

// AgentState represents the current state of an agent
type AgentState struct {
	Name     string     `json:"name"`
	Holdings []Position `json:"holdings"`
	Stats    AgentStats `json:"stats"`
	Mu       sync.Mutex
}

// Agent is the interface that all trading strategists must implement.
// It defines the core capabilities required for participation in the simulated market.
type Agent interface {
	// GetName returns the unique name of the agent (e.g., "RNG_Strategist_1").
	GetName() string

	// Run starts the agent's primary trading loop.
	// It should block until the context is cancelled or a critical error occurs.
	// Context allows for graceful shutdown.
	Run(ctx context.Context) error

	// Stop initiates a graceful shutdown of the agent.
	// Implementations should handle cleanup, saving final state, etc.
	// This might internally cancel the context passed to Run or use another mechanism.
	Stop(ctx context.Context) error

	// GetHoldings retrieves the agent's current stock positions.
	// Returns a slice of Position or an error.
	GetHoldings(ctx context.Context) ([]Position, error)

	// GetCashBalance retrieves the agent's current available cash balance.
	GetCashBalance(ctx context.Context) (decimal.Decimal, error)

	// GetCurrentPortfolioValue calculates the total value of the agent's portfolio
	// (cash + market value of all holdings).
	GetCurrentPortfolioValue(ctx context.Context) (decimal.Decimal, error)

	// GetStats returns the current aggregated performance statistics for the agent.
	GetStats(ctx context.Context) (AgentStats, error)

	// SaveStats persists the current state and performance metrics of the agent.
	// This is crucial for long-running agents and for providing historical data
	// to the observability dashboard.
	SaveState(ctx context.Context) error

	// LoadState loads the previous state of the agent (e.g., from a file or DB).
	// This is useful if agents need to resume operations or track long-term performance.
	LoadState(ctx context.Context) error

	SetBroker(broker Broker)
}

// Broker defines the interface for interacting with the trading broker.
type Broker interface {
	SubmitTrade(ctx context.Context, trade *Trade)
}
