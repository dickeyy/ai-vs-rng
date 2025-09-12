# CIS 320 Project

_Project name TBD_

## Project Description

This project explores the efficacy of various trading strategies in a simulated stock market environment. The core objective is to compare the performance of different approaches, pure randomness, an autonomously prompted Large Language Model (LLM), and a human-input augmented LLM, against each other using fictional capital. Through this comparative analysis, the project aims to shed light on the potential and limitations of AI-driven decision-making in financial markets, as well as the impact of human intuition and guidance.

## Members

- Maya Rodriguez - Group Leader
- Kyle Dickey
- Jesus Giron
- Ana Destasio
- Kevin Chisanga
- Sebastian Dawkins

## Tech Stack

- Main language: [Go v1.24.1](https://go.dev/)
- Stock market API: [Alpaca](https://alpaca.markets/)
  - SDK: [alpaca-trade-api-go](https://github.com/alpacahq/alpaca-trade-api-go/)
- Database: Postgres + Redis
- LLM: TBD

## Deliverables

_tbd_

## Installation

1. Prerequisites

- Docker and Docker Compose
- Optional (for local, non-Docker runs): Go 1.24+

2. Clone the repository

```bash
git clone https://github.com/dickeyy/cis-320.git
cd cis-320
```

3. Create an `.env.local` in the project root for secrets

The program loads `./.env.local` on startup and will exit if it is missing.

- For production (everything in Docker): include only Alpaca values (DB/Redis are injected by Compose).
- For development (Docker infra + local app): include DB and Redis too so `go run` can connect to the dev containers on localhost.

```bash
# Required for production runs (in Docker) and optional for dev
ALPACA_KEY=your_key
ALPACA_SECRET=your_secret
# Optional: defaults to paper trading
ALPACA_API=https://paper-api.alpaca.markets

# Only needed when running the app locally (dev)
DATABASE_URL=postgres://postgres:postgres@localhost:5432/cis320?sslmode=disable
REDIS_URL=redis://localhost:6379/0
```

## Usage

### Development workflow (Docker infra + local app)

1. Start only Redis and Postgres using the dev compose file:

```bash
docker compose -f docker-compose.dev.yml up -d
```

2. Run the app locally with flags as needed (Alpaca creds not required in `--dev`):

```bash
go run . --debug --dev
```

3. Stop the dev infrastructure when done:

```bash
docker compose -f docker-compose.dev.yml down
```

### Production workflow (everything in Docker)

1. Ensure `.env.local` contains your Alpaca credentials.

2. Build and start the full stack (app + Redis + Postgres):

```bash
docker compose up --build -d
```

3. View logs and stop:

```bash
docker compose logs -f app
docker compose down
```

Available agents:

- `random`
- `llm`
- `llm-human`

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
