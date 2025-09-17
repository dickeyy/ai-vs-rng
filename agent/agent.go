package agent

import (
	"context"
	"time"

	"github.com/dickeyy/cis-320/types"
	"github.com/dickeyy/cis-320/utils"
	"github.com/rs/zerolog/log"
)

func StartAgents(agents []types.Agent) {
	log.Info().Msg("Starting agents")

	// Create a single centralized ticker
	var interval time.Duration = 15 * time.Minute
	if utils.DevMode {
		interval = 20 * time.Second
	}
	ticker := time.NewTicker(interval)

	// Create per-agent tick channels and set them on agents
	tickChans := make([]chan time.Time, 0, len(agents))
	for _, a := range agents {
		ch := make(chan time.Time, 1)
		tickChans = append(tickChans, ch)
		a.SetTickChannel(ch)
	}

	// Broadcast ticks to all agents
	go func() {
		for t := range ticker.C {
			for _, ch := range tickChans {
				select {
				case ch <- t:
					// delivered
					log.Debug().Msg("Tick delivered to agents")
				default:
					// drop if receiver is slow to avoid blocking
					log.Warn().Msg("Tick dropped from agents")
				}
			}
		}
	}()

	// Start each agent in its own goroutine
	for _, a := range agents {
		ag := a // create a new variable for the goroutine
		go func() {
			log.Info().Str("agent", ag.GetName()).Msg("Starting agent")
			err := ag.Run(context.Background())
			if err != nil {
				log.Err(err).Str("agent", ag.GetName()).Msg("Agent run failed")
			}
			log.Info().Str("agent", ag.GetName()).Msg("Agent stopped")
		}()
	}

	// The main function will handle the program's lifecycle and graceful shutdown.
	// No need for a ticker here as agents manage their own timing.
}
