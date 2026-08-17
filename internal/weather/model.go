package weather

import "time"

type Forecast struct {
	Latitude  float64
	Longitude float64
	Timezone  string
	Daily     []DailyForecast
}

type DailyForecast struct {
	Date             time.Time
	TemperatureMax   float64
	TemperatureMin   float64
	PrecipitationSum float64
}

type ForecastParams struct {
	Latitude     float64
	Longitude    float64
	ForecastDays int
}
