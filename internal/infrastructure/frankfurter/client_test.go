package frankfurter

import (
	"context"
	"encoding/json"
	nethttp "net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestClient(t *testing.T, baseURL string) *Client {
	t.Helper()
	return NewClient(baseURL)
}

func TestGetRate_Path(t *testing.T) {
	var capturedPath string

	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(rateDTO{
			Date: "2026-08-17", Base: "EUR", Quote: "USD", Rate: 1.1587,
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	rate, err := client.GetRate(context.Background(), "EUR", "USD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedPath != "/rate/EUR/USD" {
		t.Errorf("expected /rate/EUR/USD, got %s", capturedPath)
	}
	if rate.Rate != 1.1587 {
		t.Errorf("expected rate 1.1587, got %f", rate.Rate)
	}
	if rate.Base != "EUR" {
		t.Errorf("expected base EUR, got %s", rate.Base)
	}
	if rate.Quote != "USD" {
		t.Errorf("expected quote USD, got %s", rate.Quote)
	}
	expectedDate, _ := time.Parse("2006-01-02", "2026-08-17")
	if !rate.Date.Equal(expectedDate) {
		t.Errorf("expected date 2026-08-17, got %v", rate.Date)
	}
}

func TestGetRates_QueryParams(t *testing.T) {
	var capturedPath string
	var capturedBase, capturedQuotes string

	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		capturedPath = r.URL.Path
		capturedBase = r.URL.Query().Get("base")
		capturedQuotes = r.URL.Query().Get("quotes")

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]rateDTO{
			{Date: "2026-08-17", Base: "EUR", Quote: "USD", Rate: 1.1587},
			{Date: "2026-08-17", Base: "EUR", Quote: "GBP", Rate: 0.8556},
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	rates, err := client.GetRates(context.Background(), "EUR", []string{"USD", "GBP"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedPath != "/rates" {
		t.Errorf("expected /rates, got %s", capturedPath)
	}
	if capturedBase != "EUR" {
		t.Errorf("expected base=EUR, got %s", capturedBase)
	}
	if capturedQuotes != "USD,GBP" {
		t.Errorf("expected quotes=USD,GBP, got %s", capturedQuotes)
	}
	if len(rates) != 2 {
		t.Fatalf("expected 2 rates, got %d", len(rates))
	}
}

func TestListCurrencies_Path(t *testing.T) {
	var capturedPath string

	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]currencyDTO{
			{ISOCode: "EUR", Name: "Euro"},
			{ISOCode: "USD", Name: "United States Dollar"},
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	currencies, err := client.ListCurrencies(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedPath != "/currencies" {
		t.Errorf("expected /currencies, got %s", capturedPath)
	}
	if len(currencies) != 2 {
		t.Fatalf("expected 2 currencies, got %d", len(currencies))
	}
	if currencies[0].Code != "EUR" {
		t.Errorf("expected code EUR, got %s", currencies[0].Code)
	}
	if currencies[0].Name != "Euro" {
		t.Errorf("expected name Euro, got %s", currencies[0].Name)
	}
	if currencies[1].Code != "USD" {
		t.Errorf("expected code USD, got %s", currencies[1].Code)
	}
}

func TestGetRate_UpstreamError(t *testing.T) {
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, _ *nethttp.Request) {
		w.WriteHeader(nethttp.StatusInternalServerError)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	_, err := client.GetRate(context.Background(), "EUR", "USD")
	if err == nil {
		t.Fatal("expected error for 500")
	}
}

func TestGetRate_MalformedJSON(t *testing.T) {
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, _ *nethttp.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"date": "2026-08-17", "base":`))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	_, err := client.GetRate(context.Background(), "EUR", "USD")
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestGetRates_EmptyList(t *testing.T) {
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, _ *nethttp.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]rateDTO{})
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	rates, err := client.GetRates(context.Background(), "EUR", []string{"USD"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rates) != 0 {
		t.Fatalf("expected 0 rates, got %d", len(rates))
	}
}

func TestListCurrencies_Mapping(t *testing.T) {
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, _ *nethttp.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]currencyDTO{
			{ISOCode: "PLN", Name: "Polish Zloty"},
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	currencies, err := client.ListCurrencies(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(currencies) != 1 {
		t.Fatalf("expected 1 currency, got %d", len(currencies))
	}
	if currencies[0].Code != "PLN" || currencies[0].Name != "Polish Zloty" {
		t.Errorf("unexpected currency: %+v", currencies[0])
	}
}

func TestGetRate_NotFound(t *testing.T) {
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, _ *nethttp.Request) {
		w.WriteHeader(nethttp.StatusNotFound)
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	_, err := client.GetRate(context.Background(), "EUR", "XYZ")
	if err == nil {
		t.Fatal("expected error for 404")
	}
}
