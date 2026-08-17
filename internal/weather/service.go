package weather

import (
	"context"
	"fmt"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Forecast(ctx context.Context, params ForecastParams) (Forecast, error) {
	if params.Latitude < -90 || params.Latitude > 90 {
		return Forecast{}, fmt.Errorf("%w: latitude must be between -90 and 90", ErrInvalidInput)
	}
	if params.Longitude < -180 || params.Longitude > 180 {
		return Forecast{}, fmt.Errorf("%w: longitude must be between -180 and 180", ErrInvalidInput)
	}

	forecastDays := params.ForecastDays
	if forecastDays == 0 {
		forecastDays = 1
	}
	if forecastDays < 1 || forecastDays > 16 {
		return Forecast{}, fmt.Errorf("%w: forecast_days must be between 1 and 16", ErrInvalidInput)
	}

	f, err := s.repo.Forecast(ctx, ForecastParams{
		Latitude:     params.Latitude,
		Longitude:    params.Longitude,
		ForecastDays: forecastDays,
	})
	if err != nil {
		return Forecast{}, fmt.Errorf("weather forecast: %w", err)
	}
	return f, nil
}
