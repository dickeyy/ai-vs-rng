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

The program loads `./.env.local` on startup and will exit if it is missing. For Docker Compose, only the Alpaca values are required here (DB/Redis are injected by Compose). For a local run without Compose, include DB and Redis as well.

```bash
# Required for both Docker and local runs
ALPACA_KEY=your_key
ALPACA_SECRET=your_secret
# Optional: defaults to paper trading
ALPACA_API=https://paper-api.alpaca.markets

# Only needed when running locally without docker-compose
DATABASE_URL=postgres://postgres:postgres@localhost:5432/cis320?sslmode=disable
REDIS_URL=redis://localhost:6379/0
```

## Usage

### Run with Docker Compose (recommended)

1. Start all services (app, Redis, Postgres):

```bash
docker compose up --build
```

2. Pass flags dynamically when needed (no defaults baked in):

```bash
docker compose run --rm app --debug --dev
```

Notes:

- The Compose file mounts `.env.local` into the container so the program can load secrets.
- Postgres and Redis are provisioned automatically; migrations and `pgcrypto` are applied on first startup.

### Run locally (without Docker)

1. Ensure Postgres and Redis are running and that `DATABASE_URL` and `REDIS_URL` are set in `.env.local` as shown above (or exported in your shell).

2. Build and run:

```bash
go run . --debug --dev
# or
go build -o cis-320 . && ./cis-320
```

3. Optionally, use Docker for infra only and run the app locally:

```bash
docker compose up -d postgres redis
go run . --debug --dev
```

Available agents:

- `random`
- `llm`
- `llm-human`

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
