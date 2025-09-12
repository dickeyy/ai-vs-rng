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
			log.Info().Msgf("Starting agent: %s", agent.GetName())
			err := agent.Run(context.Background())
			if err != nil {
				log.Err(err).Msgf("Agent %s Run failed", agent.GetName())
			}
			log.Info().Msgf("Agent %s stopped", agent.GetName())
		}()
	}

	// The main function will handle the program's lifecycle and graceful shutdown.
	// No need for a ticker here as agents manage their own timing.
}
