package frankfurter

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"mcp-example/internal/currency"
	"mcp-example/pkg/http"
)

type Client struct {
	client  *http.Client
	baseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{client: http.New(), baseURL: baseURL}
}

func (c *Client) GetRate(ctx context.Context, base, quote string) (currency.Rate, error) {
	u := fmt.Sprintf("%s/rate/%s/%s", c.baseURL, base, quote)

	var dto rateDTO
	if err := c.client.DoJSON(ctx, "GET", u, &dto); err != nil {
		return currency.Rate{}, mapError(err)
	}

	return mapRate(dto)
}

func (c *Client) GetRates(ctx context.Context, base string, quotes []string) ([]currency.Rate, error) {
	u := fmt.Sprintf("%s/rates?base=%s&quotes=%s", c.baseURL, base, strings.Join(quotes, ","))

	var dtos []rateDTO
	if err := c.client.DoJSON(ctx, "GET", u, &dtos); err != nil {
		return nil, mapError(err)
	}

	rates := make([]currency.Rate, 0, len(dtos))
	for _, dto := range dtos {
		r, err := mapRate(dto)
		if err != nil {
			return nil, err
		}
		rates = append(rates, r)
	}
	return rates, nil
}

func (c *Client) ListCurrencies(ctx context.Context) ([]currency.Currency, error) {
	u := fmt.Sprintf("%s/currencies", c.baseURL)

	var dtos []currencyDTO
	if err := c.client.DoJSON(ctx, "GET", u, &dtos); err != nil {
		return nil, mapError(err)
	}

	currencies := make([]currency.Currency, 0, len(dtos))
	for _, dto := range dtos {
		currencies = append(currencies, currency.Currency{
			Code: dto.ISOCode,
			Name: dto.Name,
		})
	}
	return currencies, nil
}

func mapRate(dto rateDTO) (currency.Rate, error) {
	date, err := time.Parse("2006-01-02", dto.Date)
	if err != nil {
		return currency.Rate{}, fmt.Errorf("parse rate date %q: %w", dto.Date, err)
	}
	return currency.Rate{
		Date:  date,
		Base:  dto.Base,
		Quote: dto.Quote,
		Rate:  dto.Rate,
	}, nil
}

func mapError(err error) error {
	var httpErr *http.HTTPError
	if errors.As(err, &httpErr) {
		if httpErr.IsNotFound() {
			return fmt.Errorf("%w: %s", currency.ErrNotFound, httpErr.Status)
		}
		return fmt.Errorf("%w: %s", currency.ErrUpstream, httpErr.Status)
	}
	return err
}
