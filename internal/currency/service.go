package currency

import (
	"context"
	"fmt"
	"strings"
	"unicode"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetRate(ctx context.Context, base, quote string) (Rate, error) {
	b := normalizeCode(base)
	q := normalizeCode(quote)

	if err := validateCode(b, "base"); err != nil {
		return Rate{}, err
	}
	if err := validateCode(q, "quote"); err != nil {
		return Rate{}, err
	}
	if b == q {
		return Rate{}, fmt.Errorf("%w: base and quote must differ", ErrInvalidInput)
	}

	rate, err := s.repo.GetRate(ctx, b, q)
	if err != nil {
		return Rate{}, fmt.Errorf("get rate %s/%s: %w", b, q, err)
	}
	return rate, nil
}

func (s *Service) GetRates(ctx context.Context, base string, quotes []string) ([]Rate, error) {
	b := normalizeCode(base)
	if err := validateCode(b, "base"); err != nil {
		return nil, err
	}

	if len(quotes) < 1 || len(quotes) > 20 {
		return nil, fmt.Errorf("%w: must provide between 1 and 20 quote currencies", ErrInvalidInput)
	}

	normalized := make([]string, 0, len(quotes))
	for i, q := range quotes {
		nq := normalizeCode(q)
		if err := validateCode(nq, fmt.Sprintf("quote[%d]", i)); err != nil {
			return nil, err
		}
		if nq == b {
			return nil, fmt.Errorf("%w: quote currency must differ from base", ErrInvalidInput)
		}
		normalized = append(normalized, nq)
	}

	rates, err := s.repo.GetRates(ctx, b, normalized)
	if err != nil {
		return nil, fmt.Errorf("get rates for base %s: %w", b, err)
	}
	return rates, nil
}

func (s *Service) ListCurrencies(ctx context.Context) ([]Currency, error) {
	currencies, err := s.repo.ListCurrencies(ctx)
	if err != nil {
		return nil, fmt.Errorf("list currencies: %w", err)
	}
	return currencies, nil
}

func (s *Service) Convert(ctx context.Context, params ConversionParams) (Conversion, error) {
	if params.Amount <= 0 {
		return Conversion{}, fmt.Errorf("%w: amount must be greater than zero", ErrInvalidInput)
	}

	from := normalizeCode(params.From)
	to := normalizeCode(params.To)

	if err := validateCode(from, "from"); err != nil {
		return Conversion{}, err
	}
	if err := validateCode(to, "to"); err != nil {
		return Conversion{}, err
	}
	if from == to {
		return Conversion{}, fmt.Errorf("%w: from and to must differ", ErrInvalidInput)
	}

	rate, err := s.repo.GetRate(ctx, from, to)
	if err != nil {
		return Conversion{}, fmt.Errorf("convert %s to %s: %w", from, to, err)
	}

	return Conversion{
		Date:      rate.Date,
		From:      from,
		To:        to,
		Amount:    params.Amount,
		Rate:      rate.Rate,
		Converted: params.Amount * rate.Rate,
	}, nil
}

func normalizeCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

func validateCode(code, field string) error {
	if len(code) != 3 {
		return fmt.Errorf("%w: %s must be a 3-letter currency code", ErrInvalidInput, field)
	}
	for _, r := range code {
		if !unicode.IsLetter(r) {
			return fmt.Errorf("%w: %s must contain only alphabetic characters", ErrInvalidInput, field)
		}
	}
	return nil
}
