package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/dickeyy/cis-320/types"
	"github.com/dickeyy/cis-320/utils"
	openrouter "github.com/revrost/go-openrouter"
	"github.com/revrost/go-openrouter/jsonschema"
	"github.com/rs/zerolog/log"
)

var (
	AI                *openrouter.Client
	SystemPrompt      string
	PreviousResponses []string
)

func InitializeAI() {
	AI = openrouter.NewClient(os.Getenv("OPENROUTER_KEY"), openrouter.WithXTitle("CIS-320"))

	s, err := utils.GetSystemPrompt()
	if err != nil {
		log.Fatal().Err(err).Msg("Error getting system prompt")
	}
	SystemPrompt = s

	log.Info().Msg("System prompt initialized")
}

func GetAITradeDecision(ctx context.Context, agentState *types.AgentState, lastError error) (*types.TradeDecision, error) {
	userPrompt, err := utils.GetUserPrompt(agentState, PreviousResponses, lastError)
	if err != nil {
		return nil, fmt.Errorf("failed to get user prompt: %w", err)
	}

	type Result struct {
		TradeDecision types.TradeDecision `json:"trade_decision"`
	}
	var result Result
	_, err = jsonschema.GenerateSchemaForType(result)
	if err != nil {
		return nil, fmt.Errorf("failed to generate schema: %w", err)
	}

	request := openrouter.ChatCompletionRequest{
		Model: "google/gemini-2.5-flash",
		Messages: []openrouter.ChatCompletionMessage{
			{
				Role:    openrouter.ChatMessageRoleSystem,
				Content: openrouter.Content{Text: SystemPrompt},
			},
			{
				Role:    openrouter.ChatMessageRoleUser,
				Content: openrouter.Content{Text: userPrompt},
			},
		},
		// Temperature: 0.7,
	}

	res, err := AI.CreateChatCompletion(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion: %w", err)
	}

	content := res.Choices[0].Message.Content.Text
	// println(content)

	tradeDecision, err := parseTradeDecisionFromText(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse trade decision: %w", err)
	}

	PreviousResponses = append(PreviousResponses, content)

	if utils.DevMode {
		log.Debug().Any("trade_decision", tradeDecision).Msg("Trade decision")
	}

	return tradeDecision, nil
}

// parseTradeDecisionFromText extracts a JSON object from the model response text
// (supports fenced code blocks) and unmarshals it into a TradeDecision.
func parseTradeDecisionFromText(text string) (*types.TradeDecision, error) {
	// Prefer fenced code block if present
	if block, ok := extractFromCodeFence(text); ok {
		td, err := unmarshalTradeDecision(strings.TrimSpace(block))
		if err == nil {
			return td, nil
		}
		// fall through to brace-based extraction as a fallback
	}

	jsonStr, err := extractFirstJSONObject(text)
	if err != nil {
		return nil, fmt.Errorf("no JSON object found in response: %w", err)
	}
	return unmarshalTradeDecision(strings.TrimSpace(jsonStr))
}

func unmarshalTradeDecision(jsonStr string) (*types.TradeDecision, error) {
	// If top-level object includes a trade_decision key, use its value; otherwise use the object itself
	var root map[string]json.RawMessage
	var payload []byte
	if err := json.Unmarshal([]byte(jsonStr), &root); err == nil && root != nil {
		if v, ok := root["trade_decision"]; ok {
			payload = v
		} else {
			payload = []byte(jsonStr)
		}
	} else {
		payload = []byte(jsonStr)
	}

	var td types.TradeDecision
	if err := json.Unmarshal(payload, &td); err != nil {
		return nil, fmt.Errorf("invalid trade decision JSON: %w", err)
	}
	// Normalize action casing
	td.Action = strings.ToUpper(strings.TrimSpace(td.Action))
	return &td, nil
}

// extractFromCodeFence returns the contents of the first triple-backtick code block, if any.
func extractFromCodeFence(s string) (string, bool) {
	start := strings.Index(s, "```")
	if start == -1 {
		return "", false
	}
	rest := s[start+3:]
	// Optionally skip a language identifier like "json" on the first line
	if nl := strings.Index(rest, "\n"); nl != -1 {
		first := rest[:nl]
		if len(first) > 0 && !strings.Contains(first, "{") {
			rest = rest[nl+1:]
		}
	}
	end := strings.Index(rest, "```")
	if end == -1 {
		return "", false
	}
	return rest[:end], true
}

// extractFirstJSONObject attempts to locate the first top-level JSON object in text.
// It uses brace counting and is safe against braces inside string literals.
func extractFirstJSONObject(s string) (string, error) {
	s = strings.TrimSpace(s)
	start := strings.IndexByte(s, '{')
	if start == -1 {
		return "", fmt.Errorf("no opening brace found")
	}
	count := 0
	inString := false
	escape := false
	for i := start; i < len(s); i++ {
		c := s[i]
		if inString {
			if escape {
				escape = false
				continue
			}
			if c == '\\' {
				escape = true
				continue
			}
			if c == '"' {
				inString = false
			}
			continue
		}
		switch c {
		case '"':
			inString = true
		case '{':
			count++
		case '}':
			count--
			if count == 0 {
				return s[start : i+1], nil
			}
		}
	}
	return "", fmt.Errorf("unterminated JSON object")
}
