package repository

import (
	"fmt"
	"tradingViewWebhookBot/internal/domain"

	"github.com/jmoiron/sqlx"
)

type CoinRepository struct {
	db *sqlx.DB
}

func NewCoinRepository(db *sqlx.DB) *CoinRepository {
	return &CoinRepository{db: db}
}

func (r *CoinRepository) FindBySymbol(symbol string) (*domain.Coin, error) {
	coin := &domain.Coin{}
	query := `SELECT id, coin_name, symbol FROM coins WHERE symbol = $1`
	err := r.db.Get(coin, query, symbol)
	if err != nil {
		return nil, fmt.Errorf("coin not found with symbol: %s", symbol)
	}
	return coin, nil
}

func (r *CoinRepository) FindById(id int64) (*domain.Coin, error) {
	coin := &domain.Coin{}
	query := `SELECT id, coin_name, symbol FROM coins WHERE id = $1`
	err := r.db.Get(coin, query, id)
	if err != nil {
		return nil, fmt.Errorf("coin not found with id: %d", id)
	}
	return coin, nil
}
