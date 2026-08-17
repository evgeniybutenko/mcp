package mcp

import (
	"testing"
	"time"

	"mcp-example/internal/currency"
	"mcp-example/internal/todo"
	"mcp-example/internal/weather"
)

func TestToTodoOutput(t *testing.T) {
	td := todo.Todo{ID: 42, Text: "buy milk", Completed: true, UserID: 7}
	out := toTodoOutput(td)

	if out.ID != 42 {
		t.Errorf("expected ID 42, got %d", out.ID)
	}
	if out.Text != "buy milk" {
		t.Errorf("expected text 'buy milk', got %s", out.Text)
	}
	if !out.Completed {
		t.Error("expected completed true")
	}
	if out.UserID != 7 {
		t.Errorf("expected user_id 7, got %d", out.UserID)
	}
}

func TestToTodosListOutput(t *testing.T) {
	page := todo.TodoPage{
		Items: []todo.Todo{
			{ID: 1, Text: "a", Completed: false, UserID: 1},
			{ID: 2, Text: "b", Completed: true, UserID: 2},
		},
		Total: 2, Skip: 0, Limit: 20,
	}
	out := toTodosListOutput(page)

	if len(out.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(out.Items))
	}
	if out.Total != 2 {
		t.Errorf("expected total 2, got %d", out.Total)
	}
	if out.Limit != 20 {
		t.Errorf("expected limit 20, got %d", out.Limit)
	}
	if out.Items[1].Completed != true {
		t.Error("expected item 1 completed")
	}
}

func TestToTodosListOutput_Empty(t *testing.T) {
	page := todo.TodoPage{Items: []todo.Todo{}}
	out := toTodosListOutput(page)

	if len(out.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(out.Items))
	}
}

func TestToWeatherOutput(t *testing.T) {
	date := time.Date(2026, 8, 17, 0, 0, 0, 0, time.UTC)
	f := weather.Forecast{
		Latitude:  52.23,
		Longitude: 21.01,
		Timezone:  "GMT",
		Daily: []weather.DailyForecast{
			{Date: date, TemperatureMax: 23.9, TemperatureMin: 15.7, PrecipitationSum: 9.0},
		},
	}
	out := toWeatherOutput(f)

	if out.Latitude != 52.23 {
		t.Errorf("expected latitude 52.23, got %f", out.Latitude)
	}
	if out.Timezone != "GMT" {
		t.Errorf("expected timezone GMT, got %s", out.Timezone)
	}
	if len(out.Daily) != 1 {
		t.Fatalf("expected 1 daily, got %d", len(out.Daily))
	}
	if out.Daily[0].Date != "2026-08-17" {
		t.Errorf("expected date 2026-08-17, got %s", out.Daily[0].Date)
	}
	if out.Daily[0].TemperatureMax != 23.9 {
		t.Errorf("expected temp max 23.9, got %f", out.Daily[0].TemperatureMax)
	}
}

func TestToRateOutput(t *testing.T) {
	date := time.Date(2026, 8, 17, 0, 0, 0, 0, time.UTC)
	r := currency.Rate{Date: date, Base: "EUR", Quote: "USD", Rate: 1.17}
	out := toRateOutput(r)

	if out.Date != "2026-08-17" {
		t.Errorf("expected date 2026-08-17, got %s", out.Date)
	}
	if out.Base != "EUR" {
		t.Errorf("expected base EUR, got %s", out.Base)
	}
	if out.Quote != "USD" {
		t.Errorf("expected quote USD, got %s", out.Quote)
	}
	if out.Rate != 1.17 {
		t.Errorf("expected rate 1.17, got %f", out.Rate)
	}
}

func TestToConversionOutput(t *testing.T) {
	date := time.Date(2026, 8, 17, 0, 0, 0, 0, time.UTC)
	c := currency.Conversion{
		Date: date, From: "EUR", To: "USD", Amount: 100, Rate: 1.17, Converted: 117,
	}
	out := toConversionOutput(c)

	if out.From != "EUR" {
		t.Errorf("expected from EUR, got %s", out.From)
	}
	if out.To != "USD" {
		t.Errorf("expected to USD, got %s", out.To)
	}
	if out.Amount != 100 {
		t.Errorf("expected amount 100, got %f", out.Amount)
	}
	if out.Rate != 1.17 {
		t.Errorf("expected rate 1.17, got %f", out.Rate)
	}
	if out.Converted != 117 {
		t.Errorf("expected converted 117, got %f", out.Converted)
	}
}
