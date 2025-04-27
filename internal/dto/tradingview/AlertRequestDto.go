package tradingview

import (
	"fmt"
	"strconv"
	"tradingViewWebhookBot/internal/constants/futureType"
)

type AlertRequestDto struct {
	Tag          string `json:"tag" validate:"required"`
	Ticker       string `json:"ticker" validate:"required"`
	Price        string `json:"price" validate:"required"`
	Side         string `json:"side" validate:"required,oneof=buy sell"`
	Text         string `json:"text"`
	Interval     string `json:"interval"`
	PositionSize string `json:"positionSize"`
}

func (r AlertRequestDto) GetFuturesType() futureType.FuturesType {
	if r.Side == "sell" {
		return futureType.SHORT
	}
	return futureType.LONG
}

func (r AlertRequestDto) String() string {
	return fmt.Sprintf(
		"AlertRequest{tag: %s, ticker: %s, price: %s, side: %s, text: %s, interval: %s, positionSize: %s}",
		r.Tag,
		r.Ticker,
		r.Price,
		r.Side,
		r.Text,
		r.Interval,
		r.PositionSize,
	)
}

func (r AlertRequestDto) GetPriceFloat() float64 {
	price, err := strconv.ParseFloat(r.Price, 64)
	if err != nil {
		return 0
	}
	return price
}
