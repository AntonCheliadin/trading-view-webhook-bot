package order

import (
	"github.com/spf13/viper"
	"strconv"
	"time"
)

type OrderHistoryDto struct {
	RetCode int    `mapstructure:"retCode"`
	RetMsg  string `mapstructure:"retMsg"`
	Result  struct {
		Category string         `mapstructure:"category"`
		List     []OrderDetails `mapstructure:"list"`
	} `mapstructure:"result"`
	Time int64 `mapstructure:"time"`
}

type OrderDetails struct {
	AvgPrice           string `mapstructure:"avgPrice"`
	BlockTradeId       string `mapstructure:"blockTradeId"`
	CancelType         string `mapstructure:"cancelType"`
	CloseOnTrigger     bool   `mapstructure:"closeOnTrigger"`
	CreateType         string `mapstructure:"createType"`
	CreatedTime        string `mapstructure:"createdTime"`
	CumExecFee         string `mapstructure:"cumExecFee"`
	CumExecQty         string `mapstructure:"cumExecQty"`
	CumExecValue       string `mapstructure:"cumExecValue"`
	IsLeverage         string `mapstructure:"isLeverage"`
	LastPriceOnCreated string `mapstructure:"lastPriceOnCreated"`
	LeavesQty          string `mapstructure:"leavesQty"`
	LeavesValue        string `mapstructure:"leavesValue"`
	OrderId            string `mapstructure:"orderId"`
	OrderStatus        string `mapstructure:"orderStatus"`
	OrderType          string `mapstructure:"orderType"`
	Price              string `mapstructure:"price"`
	Qty                string `mapstructure:"qty"`
	Side               string `mapstructure:"side"`
	Symbol             string `mapstructure:"symbol"`
	TimeInForce        string `mapstructure:"timeInForce"`
	UpdatedTime        string `mapstructure:"updatedTime"`
}

func (d *OrderDetails) CalculateAvgPrice() float64 {
	parseFloat, _ := strconv.ParseFloat(d.AvgPrice, 64)
	return parseFloat
}

func (d *OrderDetails) CalculateTotalCost() float64 {
	return float64(d.CalculateAvgPrice()) * d.GetAmount()
}

func (d *OrderDetails) CalculateCommissionInUsd() float64 {
	return (float64(d.CalculateTotalCost()) * viper.GetFloat64("api.bybit.commission"))
}

func (d *OrderDetails) GetAmount() float64 {
	amount, _ := strconv.ParseFloat(d.Qty, 64)
	return amount
}

func (d *OrderDetails) GetCreatedAt() *time.Time {
	return nil
}
