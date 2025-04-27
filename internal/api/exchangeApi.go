package api

import (
	"time"
	"tradingViewWebhookBot/internal/constants/futureType"
	"tradingViewWebhookBot/internal/domain"
)

type ExchangeApi interface {
	//GetCurrentCoinPriceForFutures(coin *domain.Coin) (float64, error)
	GetCurrentCoinPrice(coin *domain.Coin) (float64, error)
	//GetKlines(coin *domain.Coin, interval string, limit int, fromTime time.Time) (KlinesDto, error)
	//GetKlinesFutures(coin *domain.Coin, interval string, limit int, fromTime time.Time) (KlinesDto, error)

	BuyCoinByMarket(coin *domain.Coin, amount float64, price float64) (OrderResponseDto, error)
	SellCoinByMarket(coin *domain.Coin, amount float64, price float64) (OrderResponseDto, error)

	OpenFuturesOrder(coin *domain.Coin, amount float64, price float64, futuresType futureType.FuturesType, stopLossPriceInCents float64) (OrderResponseDto, error)
	CloseFuturesOrder(coin *domain.Coin, openedTransaction *domain.Transaction, price float64) (OrderResponseDto, error)
	//IsFuturesPositionOpened(coin *domain.Coin, openedOrder *domain.Transaction) bool
	//GetCloseTradeRecord(coin *domain.Coin, openTransaction *domain.Transaction) (OrderResponseDto, error)
	//GetLastFuturesOrder(coin *domain.Coin, clientOrderId string) (OrderResponseDto, error)
	//
	GetWalletBalance() (WalletBalanceDto, error)
	SetFuturesLeverage(coin *domain.Coin, leverage int) error
	SetIsolatedMargin(coin *domain.Coin, leverage int) error
	//
	//SetApiKey(apiKey string)
	//SetSecretKey(secretKey string)
}

type OrderResponseDto interface {
	CalculateAvgPrice() float64
	CalculateTotalCost() float64
	CalculateCommissionInUsd() float64
	GetAmount() float64
	GetCreatedAt() *time.Time
}

type KlinesDto interface {
	GetKlines() []KlineDto
	String() string
}

type KlineDto interface {
	GetSymbol() string
	GetInterval() string
	GetStartAt() time.Time
	GetCloseAt() time.Time
	GetOpen() float64
	GetHigh() float64
	GetLow() float64
	GetClose() float64
}

type WalletBalanceDto interface {
	GetAvailableBalance() float64
}
