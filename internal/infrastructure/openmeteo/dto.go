package openmeteo

type forecastResponseDTO struct {
	Latitude  float64  `json:"latitude"`
	Longitude float64  `json:"longitude"`
	Timezone  string   `json:"timezone"`
	Daily     dailyDTO `json:"daily"`
}

type dailyDTO struct {
	Time             []string  `json:"time"`
	TemperatureMax   []float64 `json:"temperature_2m_max"`
	TemperatureMin   []float64 `json:"temperature_2m_min"`
	PrecipitationSum []float64 `json:"precipitation_sum"`
}
