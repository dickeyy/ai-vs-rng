# CIS 320 Project

<img src="https://waka.kyle.so/api/badge/kyle/interval:any/project:cis-320" alt="WakaTime">

_Project name TBD_

## Project Description

This project explores the efficacy of various trading strategies in a simulated stock market environment. The core objective is to compare the performance of different approaches, pure randomness and an autonomously prompted Large Language Model (LLM) against each other using fictional capital. Through this comparative analysis, the project aims to shed light on the potential and limitations of AI-driven decision-making in financial markets, using RNG as a baseline.

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
- LLM: [Gemini 2.0 Flash](https://openrouter.ai/google/gemini-2.0-flash-001)\*
  - Provider: [OpenRouter](https://openrouter.ai/)\*
  - SDK: [go-openrouter](https://github.com/reVrost/go-openrouter)\*

_\* = subject to change_

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
ALPACA_KEY=your_key
ALPACA_SECRET=your_secret
OPENROUTER_KEY=your_key
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
