package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

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
func GetUserPrompt(agentState *types.AgentState, previousResponses []string) (string, error) {
	agentState.Mu.Lock()
	defer agentState.Mu.Unlock()

	accountJSON, err := json.MarshalIndent(agentState.Account, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal account state to JSON: %w", err)
	}

	holdingsJSON, err := json.MarshalIndent(agentState.Holdings, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal holdings to JSON: %w", err)
	}

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
---
**Based on the above information and your directives, generate a single JSON object representing your optimal trading decision or no action.**`,
		string(accountJSON),
		string(holdingsJSON),
		buyingPower,
		portfolioValue,
		normalizePreviousResponses(previousResponses),
		tradableSymbols,
	)

	if err := os.WriteFile("prompts/example-user-prompt.txt", []byte(userPrompt), 0644); err != nil {
		return "", fmt.Errorf("failed to write example user prompt to file: %w", err)
	}

	return userPrompt, nil
}

func normalizePreviousResponses(previousResponses []string) string {
	if len(previousResponses) > 50 {
		previousResponses = previousResponses[len(previousResponses)-50:]
	}
	return strings.Join(previousResponses, "\n")
}
