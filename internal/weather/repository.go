package weather

import "context"

type Repository interface {
	Forecast(ctx context.Context, params ForecastParams) (Forecast, error)
}
