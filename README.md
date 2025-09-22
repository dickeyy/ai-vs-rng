# CIS 320 Project

<img src="https://waka.kyle.so/api/badge/kyle/interval:any/project:cis-320" alt="WakaTime">

_Project name TBD_

## Project Description

This project explores the efficacy of various trading strategies in a simulated stock market environment. The core objective is to compare the performance of different approaches, pure randomness and an autonomously prompted Large Language Model (LLM) against each other using fictional capital. Through this comparative analysis, the project aims to shed light on the potential and limitations of AI-driven decision-making in financial markets, using RNG as a baseline.

## Members

- Ana Destasio - Group Leader
- Maya Rodriguez
- Kyle Dickey
- Jesus Giron
- Kevin Chisanga
- Sebastian Dawkins

## Tech Stack

- Main language: [Go v1.24.1](https://go.dev/)
- Stock market API: [Alpaca](https://alpaca.markets/)
  - SDK: [alpaca-trade-api-go](https://github.com/alpacahq/alpaca-trade-api-go/)
- LLM: [Gemini 2.5 Flash](https://openrouter.ai/google/gemini-2.5-flash)\*
  - Provider: [OpenRouter](https://openrouter.ai/)
  - SDK: [go-openrouter](https://github.com/reVrost/go-openrouter)

_\* = subject to change_

### Tech Stack Rationale

- Go: Go is my (Kyle) favorite language. It is fast, reliable, easy to use, and is a great language for building modular systems such as this. It is always my go-to.
- Alpaca: Alpaca provides a wonderful paper trading API and dashboard. It also has suplemental APIs such as the Clock and Assets API. It is also free for Paper trading.
- Gemini 2.5 Flash: Gemini 2.5 Flash is part of Google's latest Gemini releases. According to OpenRouter's rankings, it is ranked #1 in Finance and Academia. It is also insanely cheap and fast, each request we make costs ~$0.006.

## Deliverables

_tbd_

## Installation

```bash
git clone https://github.com/dickeyy/cis-320.git
cd cis-320
```

3. Create an `.env.local` in the project root for secrets.

The program loads `./.env.local` on startup and will exit if it is missing.

```bash
ALPACA_KEY_RNG=your_key
ALPACA_SECRET_RNG=your_secret

ALPACA_KEY_LLM=your_key
ALPACA_SECRET_LLM=your_secret

OPENROUTER_KEY=your_key

REDIS_URL=your_url

# For logging to Axiom (optional, only if you want to send logs to Axiom)
AXIOM_TOKEN=your_token
AXIOM_DATASET=your_dataset
```

## Usage

Development:

```bash
go run . --dev
```

Production:

```bash
go build . -o cis-320
./cis-320
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
