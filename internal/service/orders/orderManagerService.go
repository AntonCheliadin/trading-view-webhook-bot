package orders

import (
	"database/sql"
	"fmt"
	"go.uber.org/zap"
	"math"
	"time"
	"tradingViewWebhookBot/internal/api"
	"tradingViewWebhookBot/internal/constants"
	"tradingViewWebhookBot/internal/constants/futureType"
	"tradingViewWebhookBot/internal/domain"
	"tradingViewWebhookBot/internal/repository"
	"tradingViewWebhookBot/internal/service/date"
	telegramApi "tradingViewWebhookBot/internal/telegram"
	"tradingViewWebhookBot/internal/util"
)

var orderManagerServiceImpl *OrderManagerService

func NewOrderManagerService(transactionRepo repository.Transaction,
	exchangeApi api.ExchangeApi,
	clock date.Clock,
	telegramClient *telegramApi.TelegramClient,
	leverage int64) *OrderManagerService {
	if orderManagerServiceImpl != nil {
		panic("Unexpected try to create second service instance")
	}
	orderManagerServiceImpl = &OrderManagerService{
		transactionRepo: transactionRepo,
		exchangeApi:     exchangeApi,
		telegramClient:  telegramClient,
		Clock:           clock,
		leverage:        leverage,
	}
	return orderManagerServiceImpl
}

type OrderManagerService struct {
	transactionRepo repository.Transaction
	exchangeApi     api.ExchangeApi
	telegramClient  *telegramApi.TelegramClient
	Clock           date.Clock
	leverage        int64
}

func (s *OrderManagerService) SetFuturesLeverage(coin *domain.Coin, leverage int) error {
	err := s.exchangeApi.SetFuturesLeverage(coin, leverage)
	if err != nil {
		return err
	}
	return nil
}

func (s *OrderManagerService) SetIsolatedMargin(coin *domain.Coin, leverage int) error {
	err := s.exchangeApi.SetIsolatedMargin(coin, leverage)
	if err != nil {
		return err
	}
	return nil
}

func (s *OrderManagerService) OpenFuturesOrderWithPercentStopLoss(tradingStrategy *domain.TradingStrategy, coin *domain.Coin, tradingKey string, futuresType futureType.FuturesType, stopLossInPercent float64) {
	currentPrice, err := s.exchangeApi.GetCurrentCoinPrice(coin)
	if err != nil {
		zap.S().Errorf("Error during GetCurrentCoinPrice at %v: %s", s.Clock.NowTime(), err.Error())
		return
	}

	stopLossPrice := util.CalculatePriceForStopLoss(currentPrice, stopLossInPercent, futuresType)

	s.OpenFuturesOrderWithFixedStopLoss(tradingStrategy, coin, tradingKey, futuresType, stopLossPrice)
}

func (s *OrderManagerService) OpenFuturesOrderWithFixedStopLoss(tradingStrategy *domain.TradingStrategy, coin *domain.Coin, tradingKey string, futuresType futureType.FuturesType, stopLossPrice float64) {
	s.openOrderWithCostAndFixedStopLossAndTakeProfit(tradingStrategy, coin, tradingKey, futuresType, stopLossPrice, 0, s.getCostOfOrder(), constants.FUTURES)
}

func (s *OrderManagerService) OpenFuturesOrderWithCostAndFixedStopLossAndTakeProfit(tradingStrategy *domain.TradingStrategy, coin *domain.Coin, tradingKey string, futuresType futureType.FuturesType, cost float64, stopLossPrice float64, profitPrice float64) {
	s.openOrderWithCostAndFixedStopLossAndTakeProfit(tradingStrategy, coin, tradingKey, futuresType, stopLossPrice, profitPrice, cost, constants.FUTURES)
}

func (s *OrderManagerService) OpenFuturesOrderWithCostAndFixedStopLoss(tradingStrategy *domain.TradingStrategy, coin *domain.Coin, tradingKey string, futuresType futureType.FuturesType, cost float64, stopLossPrice float64) {
	s.openOrderWithCostAndFixedStopLossAndTakeProfit(tradingStrategy, coin, tradingKey, futuresType, stopLossPrice, 0, cost, constants.FUTURES)
}

func (s *OrderManagerService) OpenOrderWithCost(tradingStrategy *domain.TradingStrategy, coin *domain.Coin, tradingKey string, futuresType futureType.FuturesType, cost float64, tradingType constants.TradingType) {
	s.openOrderWithCostAndFixedStopLossAndTakeProfit(tradingStrategy, coin, tradingKey, futuresType, 0, 0, cost, tradingType)
}

func (s *OrderManagerService) OpenOrderAllIn(tradingStrategy *domain.TradingStrategy, coin *domain.Coin, futuresType futureType.FuturesType) {
	s.openOrderWithCostAndFixedStopLossAndTakeProfit(tradingStrategy, coin, "", futuresType, 0, 0, s.getCostOfOrder(), constants.FUTURES)
}

func (s *OrderManagerService) openOrderWithCostAndFixedStopLossAndTakeProfit(tradingStrategy *domain.TradingStrategy, coin *domain.Coin, tradingKey string, futuresType futureType.FuturesType,
	stopLossPrice float64, takeProfitPrice float64, cost float64, tradingType constants.TradingType) {
	if stopLossPrice > 0 {
		zap.S().Debugf("stopLossPrice %.2f  [%v]", stopLossPrice, s.Clock.NowTime().Format(constants.DATE_TIME_FORMAT))
	}
	if takeProfitPrice > 0 {
		zap.S().Debugf("profitPrice %.2f  [%v]", takeProfitPrice, s.Clock.NowTime().Format(constants.DATE_TIME_FORMAT))
	}

	currentPrice, err := s.exchangeApi.GetCurrentCoinPrice(coin)
	if err != nil {
		zap.S().Errorf("Error during GetCurrentCoinPrice at %v: %s", s.Clock.NowTime(), err.Error())
		return
	}

	amountTransaction := util.CalculateAmountByPriceAndCost(currentPrice, cost)
	var orderDto api.OrderResponseDto
	if tradingType == constants.FUTURES {
		orderDto, err = s.exchangeApi.OpenFuturesOrder(coin, amountTransaction, currentPrice, futuresType, stopLossPrice)
	} else if tradingType == constants.SPOT {
		orderDto, err = s.exchangeApi.BuyCoinByMarket(coin, amountTransaction, currentPrice)
	}
	if err != nil {
		zap.S().Errorf("Error during OpenFuturesOrder: %s", err.Error())
		s.telegramClient.SendMessage(fmt.Sprintf("Error during OpenFuturesOrder: %s", err.Error()))
		return
	}

	transaction := s.createOpenTransactionByOrderResponseDto(tradingStrategy, coin, tradingKey, futuresType, orderDto, stopLossPrice, takeProfitPrice)
	if err3 := s.transactionRepo.SaveTransaction(&transaction); err3 != nil {
		zap.S().Errorf("Error during SaveTransaction: %s", err3.Error())
		return
	}

	zap.S().Infof("at %s Order opened [%s] with price %v and type [%v] (0-L, 1-S)", s.Clock.NowTime().Format(constants.DATE_TIME_FORMAT), coin.Symbol, currentPrice, futuresType)
	s.telegramClient.SendMessage(coin.Symbol + " " + transaction.String())
}

func (s *OrderManagerService) CloseFuturesOrderWithCurrentPriceWithInterval(tradingStrategy *domain.TradingStrategy, coin *domain.Coin, openTransaction *domain.Transaction, interval int) *domain.Transaction {
	currentPrice, _ := s.exchangeApi.GetCurrentCoinPrice(coin)
	return s.CloseOrder(tradingStrategy, openTransaction, coin, currentPrice, constants.FUTURES)
}

func (s *OrderManagerService) CloseOrder(tradingStrategy *domain.TradingStrategy, openTransaction *domain.Transaction, coin *domain.Coin, price float64, tradingType constants.TradingType) *domain.Transaction {
	var orderResponseDto api.OrderResponseDto
	var err error
	if tradingType == constants.SPOT {
		orderResponseDto, err = s.exchangeApi.SellCoinByMarket(coin, openTransaction.Amount, price)
	} else if tradingType == constants.FUTURES {
		orderResponseDto, err = s.exchangeApi.CloseFuturesOrder(coin, openTransaction, price)
	}
	if err != nil {
		zap.S().Errorf("Error during CloseFuturesOrder: %s", err.Error())
		s.telegramClient.SendMessage(fmt.Sprintf("Error during CloseFuturesOrder: %s", err.Error()))
		return nil
	}

	closeTransaction := s.createCloseTransactionByOrderResponseDto(tradingStrategy, coin, openTransaction, orderResponseDto)
	if errT := s.transactionRepo.SaveTransaction(closeTransaction); errT != nil {
		zap.S().Errorf("Error during SaveTransaction: %s", errT.Error())
		return nil
	}

	openTransaction.RelatedTransactionId = sql.NullInt64{Int64: closeTransaction.Id, Valid: true}
	_ = s.transactionRepo.SaveTransaction(openTransaction)
	s.telegramClient.SendMessage(coin.Symbol + " " + closeTransaction.String())

	return closeTransaction
}

func (s *OrderManagerService) createOpenTransactionByOrderResponseDto(
	tradingStrategy *domain.TradingStrategy, coin *domain.Coin, tradingKey string, futuresType futureType.FuturesType,
	orderDto api.OrderResponseDto, stopLossPrice float64, takeProfitPrice float64) domain.Transaction {

	var createdAt time.Time
	if orderDto.GetCreatedAt() != nil {
		createdAt = *orderDto.GetCreatedAt()
	} else {
		createdAt = s.Clock.NowTime().Add(time.Millisecond)
	}

	transaction := domain.Transaction{
		TradingKey:        tradingKey,
		TradingStrategyId: sql.NullInt64{Int64: tradingStrategy.Id, Valid: true},
		FuturesType:       futuresType,
		CoinId:            coin.Id,
		Amount:            orderDto.GetAmount(),
		Price:             orderDto.CalculateAvgPrice(),
		TotalCost:         orderDto.CalculateTotalCost(),
		Commission:        orderDto.CalculateCommissionInUsd(),
		CreatedAt:         createdAt,
	}

	if futuresType == futureType.LONG {
		transaction.TransactionType = constants.BUY
	} else {
		transaction.TransactionType = constants.SELL
	}
	if stopLossPrice > 0 {
		transaction.StopLossPrice = sql.NullFloat64{Float64: stopLossPrice, Valid: true}
	}
	if takeProfitPrice > 0 {
		transaction.TakeProfitPrice = sql.NullFloat64{Float64: takeProfitPrice, Valid: true}
	}
	return transaction
}

func (s *OrderManagerService) createCloseTransactionByOrderResponseDto(tradingStrategy *domain.TradingStrategy,
	coin *domain.Coin, openedTransaction *domain.Transaction, orderDto api.OrderResponseDto) *domain.Transaction {

	var buyCost float64
	var sellCost float64
	var transactionType constants.TransactionType

	if openedTransaction.FuturesType == futureType.LONG {
		buyCost = openedTransaction.TotalCost
		sellCost = orderDto.CalculateTotalCost()
		transactionType = constants.SELL
	} else {
		buyCost = orderDto.CalculateTotalCost()
		sellCost = openedTransaction.TotalCost
		transactionType = constants.BUY
	}

	profitInUsd := sellCost - buyCost - orderDto.CalculateCommissionInUsd() - openedTransaction.Commission

	var createdAt time.Time
	if orderDto.GetCreatedAt() != nil {
		createdAt = *orderDto.GetCreatedAt()
	} else {
		createdAt = s.Clock.NowTime()
	}

	percentProfit := float64(profitInUsd) / float64(openedTransaction.TotalCost) * 100

	transaction := domain.Transaction{
		TradingKey:           openedTransaction.TradingKey,
		TradingStrategyId:    sql.NullInt64{Int64: tradingStrategy.Id, Valid: true},
		FuturesType:          openedTransaction.FuturesType,
		TransactionType:      transactionType,
		CoinId:               coin.Id,
		Amount:               orderDto.GetAmount(),
		Price:                orderDto.CalculateAvgPrice(),
		TotalCost:            orderDto.CalculateTotalCost(),
		Commission:           orderDto.CalculateCommissionInUsd(),
		RelatedTransactionId: sql.NullInt64{Int64: openedTransaction.Id, Valid: true},
		Profit:               sql.NullInt64{Int64: util.GetCents(profitInUsd), Valid: true},
		PercentProfit:        sql.NullFloat64{Float64: math.Round(percentProfit*100) / 100, Valid: true},
		CreatedAt:            createdAt,
		IsFake:               openedTransaction.IsFake,
	}
	return &transaction
}

func (s *OrderManagerService) getCostOfOrder() float64 {
	walletBalanceDto, err := s.exchangeApi.GetWalletBalance()
	if err != nil {
		zap.S().Errorf("Error during GetWalletBalance at %v: %s", s.Clock.NowTime(), err.Error())
		s.telegramClient.SendMessage(fmt.Sprintf("Error getting wallet balance: %s", err.Error()))
		return 0
	}

	maxOrderCost := (walletBalanceDto.GetAvailableBalance() - 50) * float64(s.leverage)

	return maxOrderCost
}

func (s *OrderManagerService) CalculateCurrentProfitInPercentWithoutLeverage(coin *domain.Coin, openedTransaction *domain.Transaction) (float64, error) {
	currentPrice, err := s.exchangeApi.GetCurrentCoinPrice(coin)
	if err != nil {
		zap.S().Errorf("Error during GetCurrentCoinPrice at %v: %s", s.Clock.NowTime(), err.Error())
		return 0, err
	}

	currentProfitInPercent := util.CalculateProfitInPercent(openedTransaction.Price, currentPrice, openedTransaction.FuturesType)

	return currentProfitInPercent, nil
}

func (s *OrderManagerService) CalculateCurrentProfitInPercentWithLeverage(coin *domain.Coin, openedTransaction *domain.Transaction) (float64, error) {
	currentPrice, err := s.exchangeApi.GetCurrentCoinPrice(coin)
	if err != nil {
		zap.S().Errorf("Error during GetCurrentCoinPrice at %v: %s", s.Clock.NowTime(), err.Error())
		return 0, err
	}

	currentProfitInPercent := util.CalculateProfitInPercentWithLeverage(openedTransaction.Price, currentPrice, openedTransaction.FuturesType, s.leverage)

	return currentProfitInPercent, nil
}
