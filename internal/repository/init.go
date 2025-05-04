package repository

import (
	"github.com/jmoiron/sqlx"
	"time"
	"tradingViewWebhookBot/internal/constants"
	"tradingViewWebhookBot/internal/domain"
	"tradingViewWebhookBot/internal/dto/postgres/transaction"
)

type Coin interface {
	FindBySymbol(symbol string) (*domain.Coin, error)
	FindById(id int64) (*domain.Coin, error)
}

type Transaction interface {
	FindById(id int64) (*domain.Transaction, error)
	FindLastByCoinId(coinId int64, tradingStrategy domain.TradingStrategy) (*domain.Transaction, error)
	FindLastByCoinIdAndType(coinId int64, transactionType constants.TransactionType, tradingStrategy domain.TradingStrategy) (*domain.Transaction, error)
	FindLastBoughtNotSold(coinId int64, tradingStrategy domain.TradingStrategy) (*domain.Transaction, error)
	FindLastBoughtNotSoldAndDate(date time.Time, tradingStrategy domain.TradingStrategy) (*domain.Transaction, error)
	SaveTransaction(transaction *domain.Transaction) error
	CalculateSumOfProfit(tradingStrategy domain.TradingStrategy) (int64, error)
	CalculateSumOfProfitByCoin(coinId int64, tradingStrategy domain.TradingStrategy) (int64, error)
	CalculateSumOfProfitByCoinAndTradingKey(coinId int64, tradingStrategy domain.TradingStrategy, tradingKey string) (int64, error)
	CalculateSumOfSpentTransactions(tradingStrategy domain.TradingStrategy) (int64, error)
	CalculateSumOfSpentTransactionsAndCreatedAfter(date time.Time, tradingStrategy domain.TradingStrategy) (int64, error)
	CalculateSumOfProfitByDate(date time.Time, tradingStrategy domain.TradingStrategy) (int64, error)
	FindMinPriceByDate(date time.Time, tradingStrategy domain.TradingStrategy) (int64, error)
	CalculateSumOfSpentTransactionsByDate(date time.Time, tradingStrategy domain.TradingStrategy) (int64, error)
	CalculateSumOfTransactionsByDateAndType(date time.Time, transType constants.TransactionType, tradingStrategy domain.TradingStrategy) (int64, error)

	FindOpenedTransaction(tradingStrategy domain.TradingStrategy) (*domain.Transaction, error)
	FindAllOpenedTransactions(tradingStrategy domain.TradingStrategy) ([]*domain.Transaction, error)
	FindOpenedTransactionByCoin(tradingStrategyId int64, coinId int64) (*domain.Transaction, error)
	FindOpenedTransactionByCoinAndTradingKey(tradingStrategy domain.TradingStrategy, coinId int64, tradingKey string) (*domain.Transaction, error)

	FindAllProfitPercents(tradingStrategy int) ([]transaction.TransactionProfitPercentsDto, error)
	FetchStatisticByDays(tradingStrategy int, coinIds []int64) ([]transaction.PairTransactionProfitPercentsDto, error)
	FindAllCoinIds(tradingStrategy int) ([]int64, error)
}

type TradingStrategy interface {
	Update(strategy *domain.TradingStrategy) error
	List() ([]domain.TradingStrategy, error)
	FindByTag(tag string) (*domain.TradingStrategy, error)
}

type Repository struct {
	Coin            Coin
	Transaction     Transaction
	TradingStrategy TradingStrategy
}

func NewRepositories(postgresDb *sqlx.DB) *Repository {
	return &Repository{
		Coin:            NewCoinRepository(postgresDb),
		Transaction:     NewTransactionRepository(postgresDb),
		TradingStrategy: NewTradingStrategyRepository(postgresDb),
	}
}
