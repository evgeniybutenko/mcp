package integration

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"mcp-example/internal/currency"
	mcpserver "mcp-example/internal/mcp"
	"mcp-example/internal/todo"
	"mcp-example/internal/weather"
)

type fakeTodoRepo struct {
	listFn    func(context.Context, todo.ListParams) (todo.TodoPage, error)
	getFn     func(context.Context, int) (todo.Todo, error)
	getByUser func(context.Context, int) (todo.TodoPage, error)
}

func (f *fakeTodoRepo) List(ctx context.Context, p todo.ListParams) (todo.TodoPage, error) {
	return f.listFn(ctx, p)
}
func (f *fakeTodoRepo) Get(ctx context.Context, id int) (todo.Todo, error) {
	return f.getFn(ctx, id)
}
func (f *fakeTodoRepo) GetByUser(ctx context.Context, uid int) (todo.TodoPage, error) {
	return f.getByUser(ctx, uid)
}

type fakeWeatherRepo struct {
	forecastFn func(context.Context, weather.ForecastParams) (weather.Forecast, error)
}

func (f *fakeWeatherRepo) Forecast(ctx context.Context, p weather.ForecastParams) (weather.Forecast, error) {
	return f.forecastFn(ctx, p)
}

type fakeCurrencyRepo struct {
	getRateFn        func(context.Context, string, string) (currency.Rate, error)
	getRatesFn       func(context.Context, string, []string) ([]currency.Rate, error)
	listCurrenciesFn func(context.Context) ([]currency.Currency, error)
}

func (f *fakeCurrencyRepo) GetRate(ctx context.Context, b, q string) (currency.Rate, error) {
	return f.getRateFn(ctx, b, q)
}
func (f *fakeCurrencyRepo) GetRates(ctx context.Context, b string, q []string) ([]currency.Rate, error) {
	return f.getRatesFn(ctx, b, q)
}
func (f *fakeCurrencyRepo) ListCurrencies(ctx context.Context) ([]currency.Currency, error) {
	return f.listCurrenciesFn(ctx)
}

func newTestServer(t *testing.T) *sdkmcp.Server {
	t.Helper()

	todoSvc := todo.NewService(&fakeTodoRepo{
		listFn: func(_ context.Context, p todo.ListParams) (todo.TodoPage, error) {
			return todo.TodoPage{
				Items: []todo.Todo{{ID: 1, Text: "test todo", Completed: false, UserID: 5}},
				Total: 1, Skip: p.Skip, Limit: p.Limit,
			}, nil
		},
		getFn: func(_ context.Context, id int) (todo.Todo, error) {
			return todo.Todo{ID: id, Text: "fetched todo", Completed: true, UserID: 10}, nil
		},
		getByUser: func(_ context.Context, uid int) (todo.TodoPage, error) {
			return todo.TodoPage{
				Items: []todo.Todo{{ID: 1, Text: "user todo", Completed: false, UserID: uid}},
				Total: 1, Skip: 0, Limit: 1,
			}, nil
		},
	})

	weatherSvc := weather.NewService(&fakeWeatherRepo{
		forecastFn: func(_ context.Context, p weather.ForecastParams) (weather.Forecast, error) {
			date := time.Date(2026, 8, 17, 0, 0, 0, 0, time.UTC)
			return weather.Forecast{
				Latitude:  p.Latitude,
				Longitude: p.Longitude,
				Timezone:  "GMT",
				Daily: []weather.DailyForecast{
					{Date: date, TemperatureMax: 23.9, TemperatureMin: 15.7, PrecipitationSum: 9.0},
				},
			}, nil
		},
	})

	currencySvc := currency.NewService(&fakeCurrencyRepo{
		getRateFn: func(_ context.Context, b, q string) (currency.Rate, error) {
			return currency.Rate{Date: time.Date(2026, 8, 17, 0, 0, 0, 0, time.UTC), Base: b, Quote: q, Rate: 1.17}, nil
		},
		getRatesFn: func(_ context.Context, b string, quotes []string) ([]currency.Rate, error) {
			rates := make([]currency.Rate, 0, len(quotes))
			for _, q := range quotes {
				rates = append(rates, currency.Rate{Date: time.Date(2026, 8, 17, 0, 0, 0, 0, time.UTC), Base: b, Quote: q, Rate: 1.0})
			}
			return rates, nil
		},
		listCurrenciesFn: func(context.Context) ([]currency.Currency, error) {
			return []currency.Currency{{Code: "EUR", Name: "Euro"}, {Code: "USD", Name: "US Dollar"}}, nil
		},
	})

	return mcpserver.NewServer(mcpserver.ServerDeps{
		TodoService:     todoSvc,
		WeatherService:  weatherSvc,
		CurrencyService: currencySvc,
	})
}

func TestToolDiscovery(t *testing.T) {
	ctx := context.Background()
	server := newTestServer(t)

	t1, t2 := sdkmcp.NewInMemoryTransports()
	if _, err := server.Connect(ctx, t1, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}

	client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: "test-client", Version: "v1.0.0"}, nil)
	session, err := client.Connect(ctx, t2, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer func() { _ = session.Close() }()

	result, err := session.ListTools(ctx, &sdkmcp.ListToolsParams{})
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}

	wantTools := []string{
		"todos_list", "todos_get", "todos_by_user",
		"weather_forecast",
		"exchange_rate", "exchange_rates", "currencies_list", "currency_convert",
	}
	if len(result.Tools) != len(wantTools) {
		t.Fatalf("expected %d tools, got %d", len(wantTools), len(result.Tools))
	}

	gotNames := make(map[string]bool, len(result.Tools))
	for _, tool := range result.Tools {
		gotNames[tool.Name] = true
	}
	for _, name := range wantTools {
		if !gotNames[name] {
			t.Errorf("missing tool: %s", name)
		}
	}
}

func TestCallTool_TodosList(t *testing.T) {
	ctx := context.Background()
	server := newTestServer(t)

	t1, t2 := sdkmcp.NewInMemoryTransports()
	if _, err := server.Connect(ctx, t1, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}

	client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: "test-client", Version: "v1.0.0"}, nil)
	session, err := client.Connect(ctx, t2, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer func() { _ = session.Close() }()

	res, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
		Name:      "todos_list",
		Arguments: map[string]any{"limit": 10, "skip": 0},
	})
	if err != nil {
		t.Fatalf("call tool: %v", err)
	}
	if res.IsError {
		t.Fatalf("tool returned error: %v", res.Content)
	}
	if res.StructuredContent == nil {
		t.Fatal("expected structured content")
	}

	data, _ := json.Marshal(res.StructuredContent)
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal structured content: %v", err)
	}
	items, ok := out["items"].([]any)
	if !ok {
		t.Fatalf("expected items array, got %T", out["items"])
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
}

func TestCallTool_TodosGet_InvalidID(t *testing.T) {
	ctx := context.Background()
	server := newTestServer(t)

	t1, t2 := sdkmcp.NewInMemoryTransports()
	if _, err := server.Connect(ctx, t1, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}

	client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: "test-client", Version: "v1.0.0"}, nil)
	session, err := client.Connect(ctx, t2, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer func() { _ = session.Close() }()

	res, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
		Name:      "todos_get",
		Arguments: map[string]any{"id": 0},
	})
	if err != nil {
		t.Fatalf("call tool: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected error for invalid id")
	}
}

func TestCallTool_WeatherForecast(t *testing.T) {
	ctx := context.Background()
	server := newTestServer(t)

	t1, t2 := sdkmcp.NewInMemoryTransports()
	if _, err := server.Connect(ctx, t1, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}

	client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: "test-client", Version: "v1.0.0"}, nil)
	session, err := client.Connect(ctx, t2, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer func() { _ = session.Close() }()

	res, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
		Name: "weather_forecast",
		Arguments: map[string]any{
			"latitude":      52.2297,
			"longitude":     21.0122,
			"forecast_days": 1,
		},
	})
	if err != nil {
		t.Fatalf("call tool: %v", err)
	}
	if res.IsError {
		t.Fatalf("tool returned error: %v", res.Content)
	}

	data, _ := json.Marshal(res.StructuredContent)
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out["timezone"] != "GMT" {
		t.Errorf("expected timezone GMT, got %v", out["timezone"])
	}
}

func TestCallTool_CurrencyConvert(t *testing.T) {
	ctx := context.Background()
	server := newTestServer(t)

	t1, t2 := sdkmcp.NewInMemoryTransports()
	if _, err := server.Connect(ctx, t1, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}

	client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: "test-client", Version: "v1.0.0"}, nil)
	session, err := client.Connect(ctx, t2, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer func() { _ = session.Close() }()

	res, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
		Name: "currency_convert",
		Arguments: map[string]any{
			"amount": 100,
			"from":   "EUR",
			"to":     "USD",
		},
	})
	if err != nil {
		t.Fatalf("call tool: %v", err)
	}
	if res.IsError {
		t.Fatalf("tool returned error: %v", res.Content)
	}

	data, _ := json.Marshal(res.StructuredContent)
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out["from"] != "EUR" {
		t.Errorf("expected from EUR, got %v", out["from"])
	}
	if out["to"] != "USD" {
		t.Errorf("expected to USD, got %v", out["to"])
	}
	if out["rate"].(float64) != 1.17 {
		t.Errorf("expected rate 1.17, got %v", out["rate"])
	}
	if out["converted"].(float64) != 117 {
		t.Errorf("expected converted 117, got %v", out["converted"])
	}
}

func TestCallTool_CurrenciesList(t *testing.T) {
	ctx := context.Background()
	server := newTestServer(t)

	t1, t2 := sdkmcp.NewInMemoryTransports()
	if _, err := server.Connect(ctx, t1, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}

	client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: "test-client", Version: "v1.0.0"}, nil)
	session, err := client.Connect(ctx, t2, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer func() { _ = session.Close() }()

	res, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
		Name:      "currencies_list",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("call tool: %v", err)
	}
	if res.IsError {
		t.Fatalf("tool returned error: %v", res.Content)
	}

	data, _ := json.Marshal(res.StructuredContent)
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	currencies, ok := out["currencies"].([]any)
	if !ok {
		t.Fatalf("expected currencies array, got %T", out["currencies"])
	}
	if len(currencies) != 2 {
		t.Fatalf("expected 2 currencies, got %d", len(currencies))
	}
}
