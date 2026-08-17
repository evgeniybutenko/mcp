package mcp

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"mcp-example/internal/currency"
	"mcp-example/internal/todo"
	"mcp-example/internal/weather"
)

type ServerDeps struct {
	TodoService     *todo.Service
	WeatherService  *weather.Service
	CurrencyService *currency.Service
}

func NewServer(deps ServerDeps) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "mcp-public-apis",
		Version: "1.0.0",
	}, nil)

	registerTodosTools(server, deps.TodoService)
	registerWeatherTools(server, deps.WeatherService)
	registerExchangeTools(server, deps.CurrencyService)

	return server
}
