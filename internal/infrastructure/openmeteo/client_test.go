package openmeteo

import (
	"context"
	"encoding/json"
	nethttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"mcp-example/internal/weather"
)

func newTestClient(t *testing.T, baseURL string) *Client {
	t.Helper()
	return NewClient(baseURL)
}

func TestForecast_PathAndQueryParams(t *testing.T) {
	var capturedPath string
	var capturedLat, capturedLon, capturedDaily, capturedDays string

	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		capturedPath = r.URL.Path
		q := r.URL.Query()
		capturedLat = q.Get("latitude")
		capturedLon = q.Get("longitude")
		capturedDaily = q.Get("daily")
		capturedDays = q.Get("forecast_days")

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(forecastResponseDTO{
			Latitude:  52.2297,
			Longitude: 21.0122,
			Timezone:  "GMT",
			Daily: dailyDTO{
				Time:             []string{"2026-08-17", "2026-08-18"},
				TemperatureMax:   []float64{23.9, 19.4},
				TemperatureMin:   []float64{15.7, 14.5},
				PrecipitationSum: []float64{9.0, 15.0},
			},
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	f, err := client.Forecast(context.Background(), weather.ForecastParams{
		Latitude: 52.2297, Longitude: 21.0122, ForecastDays: 2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedPath != "/v1/forecast" {
		t.Errorf("expected /v1/forecast, got %s", capturedPath)
	}
	if capturedLat != "52.229700" {
		t.Errorf("expected latitude 52.229700, got %s", capturedLat)
	}
	if capturedLon != "21.012200" {
		t.Errorf("expected longitude 21.012200, got %s", capturedLon)
	}
	if capturedDaily != "temperature_2m_max,temperature_2m_min,precipitation_sum" {
		t.Errorf("expected daily params, got %s", capturedDaily)
	}
	if capturedDays != "2" {
		t.Errorf("expected forecast_days=2, got %s", capturedDays)
	}
	if len(f.Daily) != 2 {
		t.Fatalf("expected 2 daily forecasts, got %d", len(f.Daily))
	}
}

func TestForecast_ResponseMapping(t *testing.T) {
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, _ *nethttp.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(forecastResponseDTO{
			Latitude:  52.23,
			Longitude: 21.01,
			Timezone:  "Europe/Warsaw",
			Daily: dailyDTO{
				Time:             []string{"2026-08-17"},
				TemperatureMax:   []float64{23.9},
				TemperatureMin:   []float64{15.7},
				PrecipitationSum: []float64{9.0},
			},
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	f, err := client.Forecast(context.Background(), weather.ForecastParams{
		Latitude: 52.23, Longitude: 21.01, ForecastDays: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if f.Latitude != 52.23 {
		t.Errorf("expected latitude 52.23, got %f", f.Latitude)
	}
	if f.Timezone != "Europe/Warsaw" {
		t.Errorf("expected timezone Europe/Warsaw, got %s", f.Timezone)
	}
	if len(f.Daily) != 1 {
		t.Fatalf("expected 1 daily forecast, got %d", len(f.Daily))
	}
	d := f.Daily[0]
	expectedDate, _ := time.Parse("2006-01-02", "2026-08-17")
	if !d.Date.Equal(expectedDate) {
		t.Errorf("expected date 2026-08-17, got %v", d.Date)
	}
	if d.TemperatureMax != 23.9 {
		t.Errorf("expected temp max 23.9, got %f", d.TemperatureMax)
	}
	if d.TemperatureMin != 15.7 {
		t.Errorf("expected temp min 15.7, got %f", d.TemperatureMin)
	}
	if d.PrecipitationSum != 9.0 {
		t.Errorf("expected precip 9.0, got %f", d.PrecipitationSum)
	}
}

func TestForecast_MalformedJSON(t *testing.T) {
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, _ *nethttp.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"latitude": 52.2, "daily":`))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	_, err := client.Forecast(context.Background(), weather.ForecastParams{
		Latitude: 52.2, Longitude: 21.0, ForecastDays: 1,
	})
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestForecast_UpstreamError(t *testing.T) {
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, _ *nethttp.Request) {
		w.WriteHeader(nethttp.StatusInternalServerError)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	_, err := client.Forecast(context.Background(), weather.ForecastParams{
		Latitude: 52.2, Longitude: 21.0, ForecastDays: 1,
	})
	if err == nil {
		t.Fatal("expected error for 500")
	}
}

func TestForecast_IncompleteDailyArrays(t *testing.T) {
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, _ *nethttp.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(forecastResponseDTO{
			Latitude:  52.23,
			Longitude: 21.01,
			Timezone:  "GMT",
			Daily: dailyDTO{
				Time:           []string{"2026-08-17", "2026-08-18"},
				TemperatureMax: []float64{23.9},
			},
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	f, err := client.Forecast(context.Background(), weather.ForecastParams{
		Latitude: 52.23, Longitude: 21.01, ForecastDays: 2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.Daily) != 2 {
		t.Fatalf("expected 2 daily forecasts, got %d", len(f.Daily))
	}
	if f.Daily[0].TemperatureMax != 23.9 {
		t.Errorf("expected temp max 23.9, got %f", f.Daily[0].TemperatureMax)
	}
	if f.Daily[1].TemperatureMax != 0 {
		t.Errorf("expected temp max 0 for missing, got %f", f.Daily[1].TemperatureMax)
	}
}

func TestForecast_EmptyDaily(t *testing.T) {
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, _ *nethttp.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(forecastResponseDTO{
			Latitude:  52.23,
			Longitude: 21.01,
			Timezone:  "GMT",
			Daily:     dailyDTO{},
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	f, err := client.Forecast(context.Background(), weather.ForecastParams{
		Latitude: 52.23, Longitude: 21.01, ForecastDays: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.Daily) != 0 {
		t.Errorf("expected 0 daily forecasts, got %d", len(f.Daily))
	}
}
