package utils

import (
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

	accountSummary := prepAccountSummary(agentState.Account)
	holdingsString := prepHoldingsString(agentState.Holdings)

	// get the data we need
	buyingPower := agentState.Account.BuyingPower.String()
	portfolioValue := agentState.Account.PortfolioValue.String()

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
- Last trade error: %s
---
**Based on the above information and your directives, generate a single JSON object representing your optimal trading decision or no action.**`,
		accountSummary,
		holdingsString,
		buyingPower,
		portfolioValue,
		normalizePreviousResponses(previousResponses),
		formatLastError(lastError),
	)

	// write the user prompt to a file
	if DevMode {
		err := os.WriteFile("prompts/example-user-prompt.txt", []byte(userPrompt), 0644)
		if err != nil {
			return "", fmt.Errorf("failed to write user prompt to file: %w", err)
		}
	}

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
		b.WriteString(fmt.Sprintf("%d. %s: %s shares @ $%s ($%s value, P&L: $%s)\n",
			i+1,
			holding.Symbol,
			holding.QtyAvailable,
			holding.CurrentPrice,
			holding.MarketValue,
			holding.UnrealizedPL))
	}
	return b.String()
}

func prepAccountSummary(account alpaca.Account) string {
	return fmt.Sprintf(`Total Portfolio Value: $%s
Available Buying Power: $%s
Cash Balance: $%s
Long Positions Value: $%s
Short Positions Value: $%s
Account Status: %s`,
		account.PortfolioValue,
		account.BuyingPower,
		account.Cash,
		account.LongMarketValue,
		account.ShortMarketValue,
		account.Status)
}

func formatLastError(lastError error) string {
	if lastError == nil {
		return "None"
	}
	return lastError.Error()
}
