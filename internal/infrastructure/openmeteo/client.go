package openmeteo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"mcp-example/internal/weather"
	"mcp-example/pkg/http"
)

type Client struct {
	client  *http.Client
	baseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{client: http.New(), baseURL: baseURL}
}

func (c *Client) Forecast(ctx context.Context, params weather.ForecastParams) (weather.Forecast, error) {
	u := fmt.Sprintf("%s/v1/forecast?latitude=%f&longitude=%f&daily=temperature_2m_max,temperature_2m_min,precipitation_sum&forecast_days=%d",
		c.baseURL, params.Latitude, params.Longitude, params.ForecastDays)

	var dto forecastResponseDTO
	if err := c.client.DoJSON(ctx, "GET", u, &dto); err != nil {
		return weather.Forecast{}, mapError(err)
	}

	return mapForecast(dto)
}

func mapForecast(dto forecastResponseDTO) (weather.Forecast, error) {
	n := len(dto.Daily.Time)
	if n == 0 {
		return weather.Forecast{}, nil
	}

	daily := make([]weather.DailyForecast, 0, n)
	for i := 0; i < n; i++ {
		date, err := time.Parse("2006-01-02", dto.Daily.Time[i])
		if err != nil {
			return weather.Forecast{}, fmt.Errorf("parse date %q: %w", dto.Daily.Time[i], err)
		}

		df := weather.DailyForecast{
			Date: date,
		}
		if i < len(dto.Daily.TemperatureMax) {
			df.TemperatureMax = dto.Daily.TemperatureMax[i]
		}
		if i < len(dto.Daily.TemperatureMin) {
			df.TemperatureMin = dto.Daily.TemperatureMin[i]
		}
		if i < len(dto.Daily.PrecipitationSum) {
			df.PrecipitationSum = dto.Daily.PrecipitationSum[i]
		}

		daily = append(daily, df)
	}

	return weather.Forecast{
		Latitude:  dto.Latitude,
		Longitude: dto.Longitude,
		Timezone:  dto.Timezone,
		Daily:     daily,
	}, nil
}

func mapError(err error) error {
	var httpErr *http.HTTPError
	if errors.As(err, &httpErr) {
		return fmt.Errorf("%w: %s", weather.ErrUpstream, httpErr.Status)
	}
	return err
}
