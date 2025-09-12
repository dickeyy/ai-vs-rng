package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dickeyy/cis-320/services"
	"github.com/dickeyy/cis-320/types"
)

// GetAgentState retrieves the last saved state of an agent from redis
func GetAgentState(ctx context.Context, agentName string) (*types.AgentState, error) {
	key := fmt.Sprintf("agent:%s", agentName)
	d, err := services.Redis.Get(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get agent state from Redis: %w", err)
	}

	var state types.AgentState
	err = json.Unmarshal([]byte(d), &state)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal agent state from Redis: %w", err)
	}

	return &state, nil
}
