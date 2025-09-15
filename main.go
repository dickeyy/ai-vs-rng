package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/dickeyy/cis-320/agent"
	"github.com/dickeyy/cis-320/broker"
	"github.com/dickeyy/cis-320/services"
	"github.com/dickeyy/cis-320/types"
	"github.com/dickeyy/cis-320/utils"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	debug   bool = false
	devMode bool = false
)

func parseFlags() {
	d := flag.Bool("debug", false, "enable debug mode")
	dev := flag.Bool("dev", false, "enable development mode (frequent trading for testing)")
	flag.Usage = func() {
		os.Stderr.WriteString("Usage: " + os.Args[0] + " [OPTIONS] <agent_type>\n")
		os.Stderr.WriteString("Example: " + os.Args[0] + " --debug --dev\n")
		os.Stderr.WriteString("\nOptions:\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	debug = *d
	devMode = *dev
}

func init() {
	parseFlags()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug || devMode {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = zerolog.New(os.Stderr).With().Caller().Logger()
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	} else {
		log.Logger = zerolog.New(os.Stderr).With().Caller().Logger()
	}

	err := godotenv.Load(".env.local")
	if err != nil {
		log.Fatal().Msg("Error loading .env file")
	}
}

func initializeServices() {
	// pass dev mode to services for simulated execution
	utils.SetDevMode(devMode)
	services.InitializeAI()
}

func initializeAgents(tradeBroker *broker.Broker) []types.Agent {
	// parse symbols
	err := utils.ParseSymbols()
	if err != nil {
		log.Fatal().Err(err).Msg("Error parsing symbols")
	}
	log.Info().Int("symbols_count", len(utils.Symbols)).Msg("Parsed symbols")

	rngAgent := agent.NewRNGAgent("RNG_Agent", utils.Symbols)
	rngAgent.SetBroker(tradeBroker)
	agentsToStart := []types.Agent{rngAgent}
	return agentsToStart
}

func main() {
	log.Info().Msg("Starting program")
	if debug {
		log.Debug().Msg("Debug mode enabled")
	}
	if devMode {
		log.Info().Msg("Dev mode enabled")
	}

	// initialize services
	initializeServices()

	// Initialize broker
	tradeBroker := broker.NewBroker()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the broker's trade processing
	tradeBroker.ProcessTrades(ctx)

	// initialize agents and pass the broker
	agents := initializeAgents(tradeBroker)

	agent.StartAgents(agents)

	// stay alive until the program is interrupted
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)
	<-done

	log.Warn().Msg("Shutting down program")
	for _, agent := range agents {
		agent.Stop(ctx)
	}
	log.Warn().Msg("Agents stopped")
}
