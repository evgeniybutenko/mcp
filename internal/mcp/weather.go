package mcp

import (
	"context"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"mcp-example/internal/weather"
)

type weatherForecastInput struct {
	Latitude     float64 `json:"latitude" jsonschema:"latitude in decimal degrees (-90 to 90)"`
	Longitude    float64 `json:"longitude" jsonschema:"longitude in decimal degrees (-180 to 180)"`
	ForecastDays int     `json:"forecast_days" jsonschema:"number of forecast days (1-16)"`
}

type weatherForecastOutput struct {
	Latitude  float64       `json:"latitude"`
	Longitude float64       `json:"longitude"`
	Timezone  string        `json:"timezone"`
	Daily     []dailyOutput `json:"daily"`
}

type dailyOutput struct {
	Date             string  `json:"date"`
	TemperatureMax   float64 `json:"temperature_max"`
	TemperatureMin   float64 `json:"temperature_min"`
	PrecipitationSum float64 `json:"precipitation_sum"`
}

func registerWeatherTools(server *mcp.Server, svc *weather.Service) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "weather_forecast",
		Description: "Get a daily weather forecast for geographic coordinates.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in weatherForecastInput) (*mcp.CallToolResult, weatherForecastOutput, error) {
		f, err := svc.Forecast(ctx, weather.ForecastParams{
			Latitude:     in.Latitude,
			Longitude:    in.Longitude,
			ForecastDays: in.ForecastDays,
		})
		if err != nil {
			return nil, weatherForecastOutput{}, err
		}
		return nil, toWeatherOutput(f), nil
	})
}

func toWeatherOutput(f weather.Forecast) weatherForecastOutput {
	daily := make([]dailyOutput, 0, len(f.Daily))
	for _, d := range f.Daily {
		daily = append(daily, dailyOutput{
			Date:             d.Date.Format(time.DateOnly),
			TemperatureMax:   d.TemperatureMax,
			TemperatureMin:   d.TemperatureMin,
			PrecipitationSum: d.PrecipitationSum,
		})
	}
	return weatherForecastOutput{
		Latitude:  f.Latitude,
		Longitude: f.Longitude,
		Timezone:  f.Timezone,
		Daily:     daily,
	}
}
