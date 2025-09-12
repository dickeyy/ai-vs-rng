package agent

import (
	"context"

	"github.com/dickeyy/cis-320/types"
	"github.com/rs/zerolog/log"
)

func StartAgents(agents []types.Agent) {
	log.Info().Msg("Starting agents")

	// Start each agent in its own goroutine
	for _, a := range agents {
		agent := a // create a new variable for the goroutine
		go func() {
			log.Info().Str("agent", agent.GetName()).Msg("Starting agent")
			err := agent.Run(context.Background())
			if err != nil {
				log.Err(err).Str("agent", agent.GetName()).Msg("Agent run failed")
			}
			log.Info().Str("agent", agent.GetName()).Msg("Agent stopped")
		}()
	}

	// The main function will handle the program's lifecycle and graceful shutdown.
	// No need for a ticker here as agents manage their own timing.
}
