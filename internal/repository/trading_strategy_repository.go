package repository

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"strings"
	"time"
	"tradingViewWebhookBot/internal/domain"
)

type tradingStrategyRepository struct {
	db *sqlx.DB
}

func NewTradingStrategyRepository(db *sqlx.DB) TradingStrategy {
	return &tradingStrategyRepository{db: db}
}

func (r *tradingStrategyRepository) Update(strategy *domain.TradingStrategy) error {
	query := `
        UPDATE trading_strategies
        SET name = $1, description = $2, tag = $3, enabled = $4, updated_at = $5
        WHERE id = $6`

	result, err := r.db.Exec(query,
		strategy.Name,
		strategy.Description,
		strategy.Tag,
		strategy.Enabled,
		time.Now(),
		strategy.Id,
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *tradingStrategyRepository) List() ([]domain.TradingStrategy, error) {
	query := `
        SELECT *
        FROM trading_strategies
        ORDER BY id`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var strategies []domain.TradingStrategy
	for rows.Next() {
		var strategy domain.TradingStrategy
		err := rows.Scan(
			&strategy.Id,
			&strategy.Name,
			&strategy.Description,
			&strategy.Tag,
			&strategy.Enabled,
			&strategy.CreatedAt,
			&strategy.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		strategies = append(strategies, strategy)
	}
	return strategies, rows.Err()
}

func (r *tradingStrategyRepository) FindByTag(tag string) (*domain.TradingStrategy, error) {
	var strategy domain.TradingStrategy
	if err := r.db.Get(&strategy, "SELECT * FROM trading_strategies WHERE enabled = true AND tag = $1", tag); err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, nil
		}
		return nil, err
	}

	return &strategy, nil
}
