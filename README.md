# MCP Public APIs Server

A Go MCP (Model Context Protocol) server that exposes three public APIs as MCP tools:

- **Todos** via [DummyJSON](https://dummyjson.com)
- **Weather** via [Open-Meteo](https://open-meteo.com)
- **Currency / Exchange Rates** via [Frankfurter](https://frankfurter.dev)

Built with the official [Go MCP SDK](https://github.com/modelcontextprotocol/go-sdk).

---

## Overview

This is a Go MCP server that exposes three independent public API integrations as MCP tools. It runs over stdio transport and provides 8 tools for todo management, weather forecasts, and currency conversion.

## APIs

### DummyJSON Todos

Provides read access to the Todos resource from [DummyJSON](https://dummyjson.com/docs/todos).

### Open-Meteo

Provides daily weather forecasts from the [Open-Meteo Forecast API](https://open-meteo.com/en/docs), requesting `temperature_2m_max`, `temperature_2m_min`, and `precipitation_sum` for up to 16 days.

### Frankfurter

Provides exchange rate lookups and currency conversion via the [Frankfurter v2 API](https://frankfurter.dev). No API key required.

## Architecture

```
                         MCP Client
                             |
                           JSON-RPC
                              |
                              v
                     +-----------------+
                     |   MCP Server    |
                     +--------+--------+
                              |
              +---------------+---------------+
              |               |               |
              v               v               v
            Todos          Weather         Currency
           Service         Service          Service
              |               |               |
              v               v               v
         Repository      Repository       Repository
              ^               ^               ^
              |               |               |
              v               v               v
         DummyJSON        Open-Meteo      Frankfurter
          Adapter          Adapter          Adapter
              |               |               |
              v               v               v
         pkg/http (shared package, per-adapter instances)
```

The dependency direction is:

```
MCP -> Application/Domain -> Interfaces <- Infrastructure
```

The domain packages define repository interfaces. Infrastructure adapters implement those interfaces. The composition root (`cmd/server/main.go`) wires everything together.

## Dependency Inversion

Repository interfaces live in the domain packages (`internal/todo`, `internal/weather`, `internal/currency`), not in infrastructure. The infrastructure adapters (`dummyjson/todos`, `openmeteo`, `frankfurter`) implement these interfaces.

```
Domain
  |
  | defines what it needs
  v
Repository interface
  ^
  |
  | implements
  |
Infrastructure adapter
```

For example:

```
currency
    |
    | Repository (GetRate, GetRates, ListCurrencies)
    v
FrankfurterClient
```

If Frankfurter is replaced with another provider, the currency domain package requires no changes — only a new adapter implementing `currency.Repository` and a one-line change in the composition root.

## DTO mapping

Each external API has its own DTO structures in the infrastructure layer. DTOs are mapped to domain models before crossing the boundary. Domain models contain no JSON tags or provider-specific structures.

```
Open-Meteo JSON
      |
      v
openmeteo DTO
      |
      v
mapper
      |
      v
weather.Forecast
```

The same principle applies to DummyJSON and Frankfurter.

## MCP tool discovery

Tool discovery (`tools/list`) is handled by the MCP SDK. The application registers tools via `mcp.AddTool` and the SDK handles the protocol. There is no manual `tools/list` implementation.

## Transport

The server uses **stdio** transport:

```
stdin  -> MCP requests
stdout -> MCP protocol responses
stderr -> logs
```

Stdio is the simplest transport for MCP and works naturally with CLI-based MCP clients. Logs go to stderr via `log/slog` so stdout remains clean for the MCP protocol.

## MCP tools

| Tool | Purpose |
|---|---|
| `todos_list` | List todos |
| `todos_get` | Get todo by ID |
| `todos_by_user` | Get todos for a user |
| `weather_forecast` | Get weather forecast |
| `exchange_rate` | Get one exchange rate |
| `exchange_rates` | Get multiple exchange rates |
| `currencies_list` | List currencies |
| `currency_convert` | Convert an amount |

## Testing

Tests are deterministic and do not require external APIs.

### Three testing levels

1. **Service unit tests** — handwritten fakes implementing repository interfaces. No HTTP. Tests validation, business rules, and error propagation.

2. **HTTP client tests** — use `httptest.NewServer` to simulate upstream APIs. Tests request construction, JSON decoding, DTO-to-domain mapping, and error handling.

3. **MCP integration tests** — use the MCP SDK's `NewInMemoryTransports` to connect a real MCP client to the server with fake services. Tests tool discovery (verifies 8 tools) and tool invocation.

```bash
make test          # run all tests
make test-race     # run with race detector
make integration-test  # run integration tests only
```

## Running locally

```bash
# Copy environment file and adjust if needed
cp .env.example .env

# Build the server
make build

# Run the server
make run

# Or run directly
go run ./cmd/server
```

The server reads from stdin and writes MCP responses to stdout. Logs go to stderr.

### Configuration

Environment variables (see `.env.example`):

| Variable | Default | Description |
|---|---|---|
| `DUMMYJSON_BASE_URL` | `https://dummyjson.com` | DummyJSON API base URL |
| `OPEN_METEO_BASE_URL` | `https://api.open-meteo.com` | Open-Meteo API base URL |
| `FRANKFURTER_BASE_URL` | `https://api.frankfurter.dev/v2` | Frankfurter API base URL |

