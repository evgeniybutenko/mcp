package currency

import "time"

type Rate struct {
	Date  time.Time
	Base  string
	Quote string
	Rate  float64
}

type Currency struct {
	Code string
	Name string
}

type Conversion struct {
	Date      time.Time
	From      string
	To        string
	Amount    float64
	Rate      float64
	Converted float64
}

type ConversionParams struct {
	Amount float64
	From   string
	To     string
}
