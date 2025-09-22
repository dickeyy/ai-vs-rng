package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/dickeyy/cis-320/types"
	r "github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

var (
	Redis *r.Client
)

func InitializeRedis() {
	opt, err := r.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
		log.Fatal().Err(err).Msg("Error parsing Redis URL")
	}

	Redis = r.NewClient(opt)
	_, err = Redis.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal().Err(err).Msg("Error pinging Redis")
	}

	log.Info().Msg("Redis initialized")
}

func SaveTrade(trade *types.Trade, ctx context.Context) error {
	json, err := json.Marshal(trade)
	if err != nil {
		return err
	}

	// increment the counter
	err = Redis.Incr(ctx, fmt.Sprintf("trades:%s:%s", trade.AgentName, trade.Action)).Err()
	if err != nil {
		return err
	}

	return Redis.Set(ctx, trade.ID, json, 0).Err()
}
