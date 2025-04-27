package binance

import (
	"github.com/spf13/viper"
	"strconv"
	"time"
)

type OrderResponseBinanceDto struct {
	Symbol              string `json:"symbol"`
	OrderId             int    `json:"orderId"`
	OrderListId         int    `json:"orderListId"`
	ClientOrderId       string `json:"clientOrderId"`
	TransactTime        int64  `json:"transactTime"`
	Price               string `json:"price"`
	OrigQty             string `json:"origQty"`
	ExecutedQty         string `json:"executedQty"`
	CummulativeQuoteQty string `json:"cummulativeQuoteQty"` // totalCost in USD
	Status              string `json:"status"`
	TimeInForce         string `json:"timeInForce"`
	Type                string `json:"type"`
	Side                string `json:"side"`
	Fills               []fill `json:"fills"`
}

type fill struct {
	Price           string `json:"price"`
	Qty             string `json:"qty"`
	Commission      string `json:"commission"`
	CommissionAsset string `json:"commissionAsset"`
}

func (f fill) getPrice() float64 {
	money, _ := strconv.ParseFloat(f.Price, 64)
	return money
}

func (f fill) getCommission() float64 {
	money, _ := strconv.ParseFloat(f.Commission, 64)
	return money
}

func (d OrderResponseBinanceDto) CalculateAvgPrice() float64 {
	price := float64(0)

	for _, fill := range d.Fills {
		price += fill.getPrice()
	}

	return price / float64(len(d.Fills))
}

func (d OrderResponseBinanceDto) CalculateTotalCost() float64 {
	parseFloat, _ := strconv.ParseFloat(d.CummulativeQuoteQty, 64)
	return parseFloat
}

func (d OrderResponseBinanceDto) CalculateCommissionInUsd() float64 {
	totalCommission := float64(0)

	for _, fill := range d.Fills {
		totalCommission += fill.getCommission() * viper.GetFloat64("api.bybit.commission")
	}

	return totalCommission
}

func (d OrderResponseBinanceDto) GetAmount() float64 {
	amount, _ := strconv.ParseFloat(d.ExecutedQty, 64)
	return amount
}
func (d OrderResponseBinanceDto) GetCreatedAt() *time.Time {
	return nil
}
