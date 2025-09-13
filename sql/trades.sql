CREATE TABLE IF NOT EXISTS trades (
    -- Primary Key: Unique identifier for each trade record
    id UUID PRIMARY KEY,

    -- Agent who executed the trade
    agent_name VARCHAR(255) NOT NULL, -- Name of the agent (e.g., "RNG_Strategist_1")

    -- Trade Details
    symbol VARCHAR(10) NOT NULL, -- Stock ticker (e.g., "AAPL"), 10 chars should be sufficient for most.
    quantity NUMERIC(18, 8) NOT NULL, -- Number of shares, allowing for fractional shares (8 decimal places common)
    amount NUMERIC(18, 4) NOT NULL, -- Total monetary amount of the trade (e.g., quantity * price)
    price NUMERIC(18, 4) NOT NULL, -- Price per share (e.g., $170.25, 4 decimal places often sufficient for stock prices)
    action VARCHAR(4) NOT NULL, -- "BUY" or "SELL"

    -- Metadata
    timestamp TIMESTAMPTZ NOT NULL, -- The time of the trade as recorded by the API/simulation
    alpaca_id VARCHAR(255) UNIQUE NOT NULL, -- Alpaca's order ID for this trade. Unique across all trades.
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW() -- When this record was inserted into our database
);

CREATE INDEX idx_trades_agent_name ON trades (agent_name);
CREATE INDEX idx_trades_symbol ON trades (symbol);
CREATE INDEX idx_trades_timestamp ON trades (timestamp DESC);