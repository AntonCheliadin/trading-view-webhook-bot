package transaction

import (
	"time"
)

// TransactionProfitPercentsDto represents profit percentages for transactions
type TransactionProfitPercentsDto struct {
	CreatedAt     time.Time `db:"created_at"`
	ProfitPercent float64   `db:"profit_percent"`
}

// PairTransactionProfitPercentsDto represents profit statistics grouped by date
type PairTransactionProfitPercentsDto struct {
	CreatedDate                string  `db:"created_date"`
	ProfitPercentOfPairedOrder float64 `db:"profit_percent_of_paired_order"`
	ProfitSum                  int64   `db:"profit_sum"`
	OrdersSize                 int     `db:"orders_size"`
}
