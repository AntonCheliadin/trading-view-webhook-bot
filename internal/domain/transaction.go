package domain

import (
	"database/sql"
	"fmt"
	"time"
	"tradingViewWebhookBot/internal/constants"
	"tradingViewWebhookBot/internal/constants/futureType"
	"tradingViewWebhookBot/internal/util"
)

type Transaction struct {
	Id int64

	CoinId int64 `db:"coin_id"`

	TransactionType constants.TransactionType `db:"transaction_type"`

	Amount float64

	Price float64

	StopLossPrice sql.NullFloat64 `db:"stop_loss_price"`

	TakeProfitPrice sql.NullFloat64 `db:"take_profit_price"`

	/* TotalCost=(amount * price) */
	TotalCost float64 `db:"total_cost"`

	Commission float64

	CreatedAt time.Time `db:"created_at"`

	/* External order id in Binance or Bybit for easy search */
	ClientOrderId sql.NullString `db:"client_order_id"`

	/* api error*/
	ApiError sql.NullString `db:"api_error"`

	/* SELL transaction must contain link to BUY transaction and the opposite */
	RelatedTransactionId sql.NullInt64 `db:"related_transaction_id"`

	/* SELL.TotalCost - BUY.TotalCost - 2 commissions */
	Profit sql.NullInt64

	/* (Profit)/BUY.TotalCost * 100% */
	PercentProfit sql.NullFloat64 `db:"percent_profit"`

	TradingStrategyId sql.NullInt64 `db:"trading_strategy_id"`

	FuturesType futureType.FuturesType `db:"futures_type"`

	IsFake bool `db:"fake"`

	TradingKey string `db:"trading_key"`
}

func (t *Transaction) String() string {
	desc := fmt.Sprintf("Transaction {amount: %v, price: %.2f, cost: %.2f",
		t.Amount, t.Price, t.TotalCost)

	if t.Profit.Valid {
		desc += fmt.Sprintf(", profit: %v", util.RoundCentsToUsd(t.Profit.Int64))
	}
	return desc + "}"
}
