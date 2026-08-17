package mcp

import (
	"context"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"mcp-example/internal/currency"
)

type exchangeRateInput struct {
	Base  string `json:"base" jsonschema:"base currency 3-letter code"`
	Quote string `json:"quote" jsonschema:"quote currency 3-letter code"`
}

type rateOutput struct {
	Date  string  `json:"date"`
	Base  string  `json:"base"`
	Quote string  `json:"quote"`
	Rate  float64 `json:"rate"`
}

type exchangeRatesInput struct {
	Base   string   `json:"base" jsonschema:"base currency 3-letter code"`
	Quotes []string `json:"quotes" jsonschema:"list of quote currency 3-letter codes (1-20)"`
}

type exchangeRatesOutput struct {
	Base  string       `json:"base"`
	Rates []rateOutput `json:"rates"`
}

type currenciesListInput struct{}

type currencyOutput struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type currenciesListOutput struct {
	Currencies []currencyOutput `json:"currencies"`
}

type currencyConvertInput struct {
	Amount float64 `json:"amount" jsonschema:"amount to convert (must be > 0)"`
	From   string  `json:"from" jsonschema:"source currency 3-letter code"`
	To     string  `json:"to" jsonschema:"target currency 3-letter code"`
}

type conversionOutput struct {
	Date      string  `json:"date"`
	From      string  `json:"from"`
	To        string  `json:"to"`
	Amount    float64 `json:"amount"`
	Rate      float64 `json:"rate"`
	Converted float64 `json:"converted"`
}

func registerExchangeTools(server *mcp.Server, svc *currency.Service) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "exchange_rate",
		Description: "Get the latest exchange rate for a currency pair.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in exchangeRateInput) (*mcp.CallToolResult, rateOutput, error) {
		r, err := svc.GetRate(ctx, in.Base, in.Quote)
		if err != nil {
			return nil, rateOutput{}, err
		}
		return nil, toRateOutput(r), nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "exchange_rates",
		Description: "Get exchange rates from one base currency to multiple quote currencies.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in exchangeRatesInput) (*mcp.CallToolResult, exchangeRatesOutput, error) {
		rates, err := svc.GetRates(ctx, in.Base, in.Quotes)
		if err != nil {
			return nil, exchangeRatesOutput{}, err
		}
		out := exchangeRatesOutput{Base: in.Base}
		for _, r := range rates {
			out.Rates = append(out.Rates, toRateOutput(r))
		}
		return nil, out, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "currencies_list",
		Description: "List supported currencies.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in currenciesListInput) (*mcp.CallToolResult, currenciesListOutput, error) {
		currencies, err := svc.ListCurrencies(ctx)
		if err != nil {
			return nil, currenciesListOutput{}, err
		}
		out := currenciesListOutput{}
		for _, c := range currencies {
			out.Currencies = append(out.Currencies, currencyOutput{Code: c.Code, Name: c.Name})
		}
		return nil, out, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "currency_convert",
		Description: "Convert an amount between two currencies using the latest exchange rate.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in currencyConvertInput) (*mcp.CallToolResult, conversionOutput, error) {
		conv, err := svc.Convert(ctx, currency.ConversionParams{
			Amount: in.Amount,
			From:   in.From,
			To:     in.To,
		})
		if err != nil {
			return nil, conversionOutput{}, err
		}
		return nil, toConversionOutput(conv), nil
	})
}

func toRateOutput(r currency.Rate) rateOutput {
	return rateOutput{
		Date:  r.Date.Format(time.DateOnly),
		Base:  r.Base,
		Quote: r.Quote,
		Rate:  r.Rate,
	}
}

func toConversionOutput(c currency.Conversion) conversionOutput {
	return conversionOutput{
		Date:      c.Date.Format(time.DateOnly),
		From:      c.From,
		To:        c.To,
		Amount:    c.Amount,
		Rate:      c.Rate,
		Converted: c.Converted,
	}
}
