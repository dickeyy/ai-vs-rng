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

	// Create a centralized aligned ticker (wall-clock aligned)
	var period time.Duration = 10 * time.Minute
	if utils.DevMode {
		period = 20 * time.Second
	}

	// Create per-agent tick channels and set them on agents
	tickChans := make([]chan time.Time, 0, len(agents))
	for _, a := range agents {
		ch := make(chan time.Time, 1)
		tickChans = append(tickChans, ch)
		a.SetTickChannel(ch)
	}

	// Broadcast aligned ticks to all agents
	go func() {
		for {
			now := time.Now()
			next := now.Truncate(period).Add(period)
			sleep := time.Until(next)
			if sleep > 0 {
				timer := time.NewTimer(sleep)
				<-timer.C
				timer.Stop()
			}

			t := time.Now()
			for _, ch := range tickChans {
				select {
				case ch <- t:
					log.Debug().Msg("Tick delivered to agents")
				default:
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
}
