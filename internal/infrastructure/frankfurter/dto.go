package frankfurter

type rateDTO struct {
	Date  string  `json:"date"`
	Base  string  `json:"base"`
	Quote string  `json:"quote"`
	Rate  float64 `json:"rate"`
}

type currencyDTO struct {
	ISOCode string `json:"iso_code"`
	Name    string `json:"name"`
}
