package currency

import "context"

type Repository interface {
	GetRate(ctx context.Context, base, quote string) (Rate, error)
	GetRates(ctx context.Context, base string, quotes []string) ([]Rate, error)
	ListCurrencies(ctx context.Context) ([]Currency, error)
}
