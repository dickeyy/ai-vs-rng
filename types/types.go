package types

import (
	"context"
	"time"
)

// Trade represents a single trade made by an agent
type Trade struct {
	Symbol    string    `json:"symbol"`    // stock ticker, e.g. "APPL"
	Quantity  int       `json:"quantity"`  // number of shares
	Price     float64   `json:"price"`     // price per share
	Action    string    `json:"action"`    // "BUY" or "SELL"
	Timestamp time.Time `json:"timestamp"` // time of the trade
	OrderID   string    `json:"order_id"`  // unique identifier for the trade
}

// Position represents a current position in a stock
type Position struct {
	Symbol      string  `json:"symbol"`
	Quantity    int     `json:"quantity"`
	AvgCost     float64 `json:"avg_cost"`     // average cost per share
	MarketValue float64 `json:"market_value"` // current market value of the position
}

// AgentStats holds key performance indicators for an agent
type AgentStats struct {
	CurrentBalance float64 `json:"current_balance"`
	InitialBalance float64 `json:"initial_balance"`
	ProfitLoss     float64 `json:"profit_loss"` // Overall P/L
	ROI            float64 `json:"roi"`         // Return on Investment
	TotalTrades    int     `json:"total_trades"`
	WinningTrades  int     `json:"winning_trades"`
	LosingTrades   int     `json:"losing_trades"`
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

	// PlaceOrder sends a buy or sell order to the simulated market API.
	// It takes a context for timeouts/cancellations and details of the trade.
	// Returns the executed trade details or an error.
	PlaceOrder(ctx context.Context, symbol string, quantity int, orderType string) (*Trade, error)

	// GetHoldings retrieves the agent's current stock positions.
	// Returns a slice of Position or an error.
	GetHoldings(ctx context.Context) ([]Position, error)

	// GetCashBalance retrieves the agent's current available cash balance.
	GetCashBalance(ctx context.Context) (float64, error)

	// GetCurrentPortfolioValue calculates the total value of the agent's portfolio
	// (cash + market value of all holdings).
	GetCurrentPortfolioValue(ctx context.Context) (float64, error)

	// GetStats returns the current aggregated performance statistics for the agent.
	GetStats(ctx context.Context) (AgentStats, error)

	// SaveStats persists the current state and performance metrics of the agent.
	// This is crucial for long-running agents and for providing historical data
	// to the observability dashboard.
	SaveStats(ctx context.Context) error

	// LoadState loads the previous state of the agent (e.g., from a file or DB).
	// This is useful if agents need to resume operations or track long-term performance.
	LoadState(ctx context.Context) error
}
