package weather

import (
	"context"
	"errors"
	"testing"
)

type fakeRepository struct {
	forecastFunc func(context.Context, ForecastParams) (Forecast, error)
}

func (f *fakeRepository) Forecast(ctx context.Context, params ForecastParams) (Forecast, error) {
	if f.forecastFunc == nil {
		return Forecast{}, nil
	}
	return f.forecastFunc(ctx, params)
}

func TestForecast_ValidCoordinates(t *testing.T) {
	svc := NewService(&fakeRepository{
		forecastFunc: func(_ context.Context, params ForecastParams) (Forecast, error) {
			if params.Latitude != 52.2297 {
				t.Errorf("expected latitude 52.2297, got %f", params.Latitude)
			}
			if params.Longitude != 21.0122 {
				t.Errorf("expected longitude 21.0122, got %f", params.Longitude)
			}
			if params.ForecastDays != 5 {
				t.Errorf("expected forecast_days 5, got %d", params.ForecastDays)
			}
			return Forecast{Latitude: params.Latitude, Longitude: params.Longitude, Timezone: "GMT"}, nil
		},
	})

	f, err := svc.Forecast(context.Background(), ForecastParams{
		Latitude: 52.2297, Longitude: 21.0122, ForecastDays: 5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Latitude != 52.2297 {
		t.Errorf("expected latitude 52.2297, got %f", f.Latitude)
	}
}

func TestForecast_InvalidLatitude(t *testing.T) {
	svc := NewService(&fakeRepository{})

	_, err := svc.Forecast(context.Background(), ForecastParams{Latitude: -91, Longitude: 0})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}

	_, err = svc.Forecast(context.Background(), ForecastParams{Latitude: 91, Longitude: 0})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestForecast_InvalidLongitude(t *testing.T) {
	svc := NewService(&fakeRepository{})

	_, err := svc.Forecast(context.Background(), ForecastParams{Latitude: 0, Longitude: -181})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}

	_, err = svc.Forecast(context.Background(), ForecastParams{Latitude: 0, Longitude: 181})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestForecast_InvalidForecastDays(t *testing.T) {
	svc := NewService(&fakeRepository{})

	_, err := svc.Forecast(context.Background(), ForecastParams{Latitude: 0, Longitude: 0, ForecastDays: 0})
	if err != nil {
		t.Fatalf("forecast_days=0 should default to 1, got error: %v", err)
	}

	_, err = svc.Forecast(context.Background(), ForecastParams{Latitude: 0, Longitude: 0, ForecastDays: 17})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for forecast_days>16, got %v", err)
	}

	_, err = svc.Forecast(context.Background(), ForecastParams{Latitude: 0, Longitude: 0, ForecastDays: -1})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for forecast_days<1, got %v", err)
	}
}

func TestForecast_RepoError(t *testing.T) {
	repoErr := errors.New("repo failure")
	svc := NewService(&fakeRepository{
		forecastFunc: func(context.Context, ForecastParams) (Forecast, error) {
			return Forecast{}, repoErr
		},
	})

	_, err := svc.Forecast(context.Background(), ForecastParams{Latitude: 0, Longitude: 0, ForecastDays: 1})
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}
