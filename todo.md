### RNG Agent — BUY flow completion checklist

Scope: RNG buy path across `agent/rng.go`, `broker/broker.go`, `services/alpaca.go`, `storage/trades.go`, `types/*`, and wiring in `main.go`.

#### Critical fixes (must address for correct, robust BUY behavior)

- [ ] Ensure filled trade has all canonical fields set post-fill

  - Files: `services/alpaca.go` (`PlaceOrder`)
  - Issue: In real mode, only `Price` and `Quantity` are set after fill; `Amount` is not recomputed. `agent.updateHoldings` expects all three.
  - Proposed fix: After an order fills (or partially fills), set `trade.Amount = price * qty` using `decimal` multiplication, not float math.

- [ ] Handle partial fills as success (with filled values)

  - Files: `services/alpaca.go` (`PlaceOrder`)
  - Issue: Polling loop only exits on `status == "filled"`; ignores `partially_filled` terminal-ish case close to timeout.
  - Proposed fix: Treat `partially_filled` (with non-zero `FilledQty`) as a valid outcome on timeout or after a grace window. Return trade with filled qty/avg price.

- [ ] Avoid pointer aliasing to Alpaca order fields

  - Files: `services/alpaca.go` (`PlaceOrder`)
  - Issue: Assigns `trade.Price = order.FilledAvgPrice` and `trade.Quantity = &order.FilledQty`. These may alias library-owned memory.
  - Proposed fix: Deep-copy values into new `decimal.Decimal` variables and assign their addresses to `trade.Price`/`trade.Quantity`.

- [ ] Do not submit zero/near-zero notional orders

  - Files: `agent/rng.go` (`makeRandomDecision`)
  - Issue: Random spend can be `$0.00`, causing broker/API errors.
  - Proposed fix: Enforce a minimum notional (e.g., `$1.00`) and clamp to available balance; skip trade if balance < minimum.

- [ ] Use filled cost, not requested notional, for cash updates

  - Files: `agent/rng.go` (`updateBasicStats`)
  - Issue: Cash balance is adjusted by requested `Amount` (notional), not actual `price * qty`.
  - Proposed fix: If `Price` and `Quantity` are present, compute actual fill cost; fall back to notional only if needed.

- [ ] Guard against empty `symbols` set
  - Files: `utils/parse-symbols.go`, `main.go`
  - Issue: `utils.RandomString(a.Symbols)` panics if the slice is empty.
  - Proposed fix: Trim and ignore blank lines in parser; fail fast if no symbols; skip buy if `len(symbols) == 0`.

#### Important improvements (correctness/quality)

- [ ] Seed the PRNG once to avoid deterministic behavior

  - Files: `main.go` (init) or a dedicated `utils` seeding function
  - Proposed fix: `rand.Seed(time.Now().UnixNano())` during startup.

- [ ] Avoid float math for money when generating notional

  - Files: `agent/rng.go` (`makeRandomDecision`)
  - Issue: Uses `float64` + rounding for spend; can introduce subtle rounding errors.
  - Proposed fix: Generate an integer number of cents (e.g., random in `[minCents, maxCents]`) and construct `decimal` from integer to two places.

- [ ] Weighted average cost basis for holdings on additional buys

  - Files: `agent/rng.go` (`updateHoldings`)
  - Issue: For repeat buys, `CPS` is overwritten with the last trade price; this loses cost basis info.
  - Proposed fix: Compute new average cost: `(old_qty*old_avg + new_qty*fill_price) / (old_qty + new_qty)`. Set `MarketValue = CPS * total_qty`.

- [ ] Consistent semantics for `CPS` and `MarketValue`

  - Files: `types/types.go`, `agent/rng.go`
  - Issue: `CPS` is labeled "current price per share" but set to last fill price; `MarketValue` is derived from that, not live market.
  - Proposed fix: Document as cost basis or add a separate field if you intend to track live price later.

- [ ] Capture and log `SaveState` errors

  - Files: `agent/rng.go` (broker callback)
  - Issue: `SaveState` error is ignored (`_ = a.SaveState(ctx)`).
  - Proposed fix: Log on error and consider retry/backoff to honor durability goals.

- [ ] Optional: Record Alpaca server order ID
  - Files: `services/alpaca.go` (`PlaceOrder`)
  - Issue: `trade.OrderID` is set to a client-generated UUID, not Alpaca order ID.
  - Proposed fix: Add field or update to store Alpaca `order.ID` for traceability; keep client ID separately if useful.

#### Broker, persistence, and API edge cases

- [ ] Persist failed BUY attempts (optional but recommended for auditability)

  - Files: `broker/broker.go`, `storage/trades.go`, `sql/trades.sql`
  - Issue: On error, trades are not persisted (comment says both successful and failed, but code only saves on success).
  - Proposed fix: Save failed attempts with a status or extend schema (e.g., add `status`, `error_message`).

- [ ] Market/clock checks before placing BUYs (optional)

  - Files: `services/alpaca.go` (pre-check) or agent level
  - Proposed fix: Optionally check market open or `TradingBlocked` and skip trade early to reduce noisy errors.

- [ ] Timeout/backoff tuning for order polling
  - Files: `services/alpaca.go`
  - Proposed fix: Consider exponential backoff or longer deadline during volatile periods; log last known status on timeout.

#### Validation and guardrails

- [ ] Skip BUY when `CurrentBalance <= 0` or `< min_notional`

  - Files: `agent/rng.go`
  - Proposed fix: Early return `nil` trade when insufficient funds.

- [ ] Clamp notional to available cash

  - Files: `agent/rng.go`
  - Proposed fix: Ensure `spend <= CurrentBalance` after rounding; if rounding pushes over, reduce by a cent.

- [ ] Concurrency safety for balance reads vs. callback updates
  - Files: `agent/rng.go`
  - Issue: `makeRandomDecision` reads balance without lock while callback writes it; could produce inconsistent spends.
  - Proposed fix: Read balance under the same mutex or cache balance updates atomically before decision.

#### Observability and metrics (nice-to-haves per project goals)

- [ ] Enrich logs for BUY lifecycle

  - Files: `agent/rng.go`, `broker/broker.go`, `services/alpaca.go`
  - Proposed fix: Log `symbol`, `notional`, `filled_qty`, `filled_avg_price`, `fill_cost`, and post-trade `cash`/`position`.

- [ ] Expose Prometheus counters/gauges
  - Proposed fix: Add metrics for submitted/filled/failed BUYs, cash balance, and per-symbol positions.

#### Tests to add

- [ ] Unit: `makeRandomDecision` respects min notional and balance clamps
- [ ] Unit: `updateBasicStats` uses `price*qty` when available
- [ ] Unit: `updateHoldings` computes weighted average cost correctly
- [ ] Integration (dev mode): BUY creates filled trade with `Price`, `Quantity`, `Amount`; holdings and cash update; state persisted to Redis and trade saved to DB
- [ ] Integration (real mode, mocked Alpaca): partial fill handled; `Amount` set; cash/holdings consistent
