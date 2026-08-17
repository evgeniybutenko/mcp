package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"mcp-example/internal/config"
	"mcp-example/internal/currency"
	"mcp-example/internal/infrastructure/dummyjson/todos"
	"mcp-example/internal/infrastructure/frankfurter"
	"mcp-example/internal/infrastructure/openmeteo"
	mcpserver "mcp-example/internal/mcp"
	"mcp-example/internal/todo"
	"mcp-example/internal/weather"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	cfg := config.Load()

	todoRepo := todos.NewClient(cfg.DummyJSONBaseURL)
	weatherRepo := openmeteo.NewClient(cfg.OpenMeteoBaseURL)
	currencyRepo := frankfurter.NewClient(cfg.FrankfurterBaseURL)

	todoSvc := todo.NewService(todoRepo)
	weatherSvc := weather.NewService(weatherRepo)
	currencySvc := currency.NewService(currencyRepo)

	server := mcpserver.NewServer(mcpserver.ServerDeps{
		TodoService:     todoSvc,
		WeatherService:  weatherSvc,
		CurrencyService: currencySvc,
	})

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := server.Run(ctx, &sdkmcp.StdioTransport{}); err != nil {
		slog.Error("server stopped with error", "error", err)
		os.Exit(1)
	}
}
