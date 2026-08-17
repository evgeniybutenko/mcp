package currency

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeRepository struct {
	getRateFn        func(context.Context, string, string) (Rate, error)
	getRatesFn       func(context.Context, string, []string) ([]Rate, error)
	listCurrenciesFn func(context.Context) ([]Currency, error)
}

func (f *fakeRepository) GetRate(ctx context.Context, base, quote string) (Rate, error) {
	if f.getRateFn == nil {
		return Rate{}, nil
	}
	return f.getRateFn(ctx, base, quote)
}

func (f *fakeRepository) GetRates(ctx context.Context, base string, quotes []string) ([]Rate, error) {
	if f.getRatesFn == nil {
		return nil, nil
	}
	return f.getRatesFn(ctx, base, quotes)
}

func (f *fakeRepository) ListCurrencies(ctx context.Context) ([]Currency, error) {
	if f.listCurrenciesFn == nil {
		return nil, nil
	}
	return f.listCurrenciesFn(ctx)
}

func TestGetRate_Normalization(t *testing.T) {
	svc := NewService(&fakeRepository{
		getRateFn: func(_ context.Context, base, quote string) (Rate, error) {
			if base != "EUR" {
				t.Errorf("expected base EUR, got %s", base)
			}
			if quote != "USD" {
				t.Errorf("expected quote USD, got %s", quote)
			}
			return Rate{Date: time.Now(), Base: base, Quote: quote, Rate: 1.17}, nil
		},
	})

	_, err := svc.GetRate(context.Background(), "eur", "usd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetRate_InvalidCode(t *testing.T) {
	svc := NewService(&fakeRepository{})

	_, err := svc.GetRate(context.Background(), "EU", "USD")
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}

	_, err = svc.GetRate(context.Background(), "EUR1", "USD")
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}

	_, err = svc.GetRate(context.Background(), "EUR", "US")
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestGetRate_EqualCurrencies(t *testing.T) {
	svc := NewService(&fakeRepository{})

	_, err := svc.GetRate(context.Background(), "EUR", "EUR")
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestConvert_AmountLEZero(t *testing.T) {
	svc := NewService(&fakeRepository{})

	_, err := svc.Convert(context.Background(), ConversionParams{Amount: 0, From: "EUR", To: "USD"})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for amount=0, got %v", err)
	}

	_, err = svc.Convert(context.Background(), ConversionParams{Amount: -1, From: "EUR", To: "USD"})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for amount<0, got %v", err)
	}
}

func TestConvert_Success(t *testing.T) {
	rate := 1.17
	amount := 100.0
	expected := amount * rate

	svc := NewService(&fakeRepository{
		getRateFn: func(_ context.Context, base, quote string) (Rate, error) {
			return Rate{Date: time.Now(), Base: base, Quote: quote, Rate: rate}, nil
		},
	})

	conv, err := svc.Convert(context.Background(), ConversionParams{Amount: amount, From: "eur", To: "usd"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conv.Rate != rate {
		t.Errorf("expected rate %f, got %f", rate, conv.Rate)
	}
	if conv.Converted != expected {
		t.Errorf("expected converted %f, got %f", expected, conv.Converted)
	}
	if conv.From != "EUR" {
		t.Errorf("expected normalized from EUR, got %s", conv.From)
	}
	if conv.To != "USD" {
		t.Errorf("expected normalized to USD, got %s", conv.To)
	}
}

func TestConvert_CalculationCorrectness(t *testing.T) {
	tests := []struct {
		amount float64
		rate   float64
		want   float64
	}{
		{100, 1.17, 117},
		{50, 0.8556, 42.78},
		{1, 1, 1},
	}
	for _, tt := range tests {
		svc := NewService(&fakeRepository{
			getRateFn: func(context.Context, string, string) (Rate, error) {
				return Rate{Date: time.Now(), Base: "EUR", Quote: "USD", Rate: tt.rate}, nil
			},
		})

		conv, err := svc.Convert(context.Background(), ConversionParams{Amount: tt.amount, From: "EUR", To: "USD"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if conv.Converted != tt.want {
			t.Errorf("amount=%f rate=%f: expected converted %f, got %f", tt.amount, tt.rate, tt.want, conv.Converted)
		}
	}
}

func TestConvert_RateFailure(t *testing.T) {
	repoErr := errors.New("rate fetch failed")
	svc := NewService(&fakeRepository{
		getRateFn: func(context.Context, string, string) (Rate, error) {
			return Rate{}, repoErr
		},
	})

	_, err := svc.Convert(context.Background(), ConversionParams{Amount: 100, From: "EUR", To: "USD"})
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}

func TestConvert_EqualCurrencies(t *testing.T) {
	svc := NewService(&fakeRepository{})
	_, err := svc.Convert(context.Background(), ConversionParams{Amount: 100, From: "EUR", To: "EUR"})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestGetRates_Valid(t *testing.T) {
	svc := NewService(&fakeRepository{
		getRatesFn: func(_ context.Context, base string, quotes []string) ([]Rate, error) {
			if base != "EUR" {
				t.Errorf("expected base EUR, got %s", base)
			}
			rates := make([]Rate, 0, len(quotes))
			for _, q := range quotes {
				rates = append(rates, Rate{Date: time.Now(), Base: base, Quote: q, Rate: 1.0})
			}
			return rates, nil
		},
	})

	rates, err := svc.GetRates(context.Background(), "eur", []string{"usd", "gbp", "pln"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rates) != 3 {
		t.Fatalf("expected 3 rates, got %d", len(rates))
	}
}

func TestGetRates_TooManyQuotes(t *testing.T) {
	svc := NewService(&fakeRepository{})
	quotes := make([]string, 21)
	for i := range quotes {
		quotes[i] = "USD"
	}
	_, err := svc.GetRates(context.Background(), "EUR", quotes)
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestGetRates_EmptyQuotes(t *testing.T) {
	svc := NewService(&fakeRepository{})
	_, err := svc.GetRates(context.Background(), "EUR", []string{})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}
