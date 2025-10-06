package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/dickeyy/cis-320/types"
)

var (
	systemPromptPath = "prompts/system-prompt.txt"
)

type UserPromptTemplateData struct {
	AccountJSON     string `json:"account_json"`
	HoldingsJSON    string `json:"holdings_json"`
	BuyingPower     string `json:"buying_power"`
	PortfolioValue  string `json:"portfolio_value"`
	TradableSymbols string `json:"tradable_symbols"`
}

func GetSystemPrompt() (string, error) {
	prompt, err := os.ReadFile(systemPromptPath)
	if err != nil {
		return "", err
	}
	return string(prompt), nil
}

// TODO: Test this out make sure it actually works and gives an output that the LLM can understand
func GetUserPrompt(agentState *types.AgentState, previousResponses []string, lastError error) (string, error) {
	agentState.Mu.Lock()
	defer agentState.Mu.Unlock()

	accountJSON, err := json.MarshalIndent(agentState.Account, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal account state to JSON: %w", err)
	}

	holdingsString := prepHoldingsString(agentState.Holdings)

	// get the data we need
	buyingPower := agentState.Account.BuyingPower.String()
	portfolioValue := agentState.Account.PortfolioValue.String()

	tradableSymbols := strings.Join(Symbols, ", ")

	userPrompt := fmt.Sprintf(`Analyze the current market context and your portfolio to make a trading decision.

---
**Current Portfolio State:**

**Account Summary:**
%s

**Current Holdings:**
%s

---
**Decision Parameters:**
- Available buying power: %s USD
- Current total portfolio value: %s USD
- List of your previous trades: %s
- List of tradable symbols: %s
- Last trade error: %s
---
**Based on the above information and your directives, generate a single JSON object representing your optimal trading decision or no action.**`,
		string(accountJSON),
		holdingsString,
		buyingPower,
		portfolioValue,
		normalizePreviousResponses(previousResponses),
		tradableSymbols,
		formatLastError(lastError),
	)

	return userPrompt, nil
}

func normalizePreviousResponses(previousResponses []string) string {
	if len(previousResponses) > 50 {
		previousResponses = previousResponses[len(previousResponses)-50:]
	}
	return strings.Join(previousResponses, "\n")
}

func prepHoldingsString(holdings []alpaca.Position) string {
	var b strings.Builder
	for i, holding := range holdings {
		b.WriteString(fmt.Sprintf("%d. Symbol: %s\n", i+1, holding.Symbol))
		b.WriteString(fmt.Sprintf("Quantity: %s\n", holding.QtyAvailable))
		b.WriteString(fmt.Sprintf("Market Value: %s\n", holding.MarketValue))
		b.WriteString(fmt.Sprintf("Current Price: %s\n", holding.CurrentPrice))
		b.WriteString(fmt.Sprintf("Last Day Price: %s\n", holding.LastdayPrice))
		b.WriteString(fmt.Sprintf("Change Today %% (0-1): %s\n", holding.ChangeToday))
		b.WriteString(fmt.Sprintf("Unrealized PL: %s\n", holding.UnrealizedPL))
		b.WriteString(fmt.Sprintf("Cost Basis: %s\n", holding.CostBasis))
		b.WriteString("\n")
	}
	return b.String()
}

func formatLastError(lastError error) string {
	if lastError == nil {
		return "None"
	}
	return lastError.Error()
}
