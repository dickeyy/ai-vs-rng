package main

import (
	"flag"
	"os"

	"github.com/dickeyy/cis-320/services"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	debug   bool = false
	devMode bool = false
	agent   string
)

func init() {
	err := godotenv.Load(".env.local")
	if err != nil {
		log.Fatal().Msg("Error loading .env file")
	}
}

func initLogger(debug bool) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
}

func parseFlags() {
	d := flag.Bool("debug", false, "enable debug mode")
	dev := flag.Bool("dev", false, "enable development mode (frequent trading for testing)")
	flag.Usage = func() {
		os.Stderr.WriteString("Usage: " + os.Args[0] + " [OPTIONS] <agent_type>\n")
		os.Stderr.WriteString("Example: " + os.Args[0] + " --debug --dev RNG\n")
		os.Stderr.WriteString("\nOptions:\n")
		flag.PrintDefaults()
		os.Stderr.WriteString("\nAvailable agent types: RNG, LLM_Self, LLM_Human\n")
		os.Stderr.WriteString("\nDev mode: Ignores market hours, trades every 30 seconds, higher frequency limits\n")
	}
	flag.Parse()
	debug = *d
	devMode = *dev

	args := flag.Args()
	if len(args) == 0 {
		log.Fatal().Msg("No agent specified. Use '--help' for usage.")
	}
	agent = args[0]

}

func main() {
	parseFlags()
	initLogger(debug)

	log.Info().Msg("Starting AI Stock Market Challenge")
	log.Debug().Msg("Debug mode enabled")
	log.Info().Msgf("Using agent: %s", agent)

	// Initialize services
	err := services.InitializeAlpaca()
	if err != nil {
		log.Fatal().Msgf("Error initializing Alpaca: %v", err)
	}
	log.Info().Msg("Alpaca client initialized")
	log.Info().Msgf("Using Alpaca account %+v", services.AlpacaAccount.AccountNumber)

	err = services.InitializeRedis()
	if err != nil {
		log.Fatal().Msgf("Error initializing Redis: %v", err)
	}
	log.Info().Msg("Redis client initialized")
}
