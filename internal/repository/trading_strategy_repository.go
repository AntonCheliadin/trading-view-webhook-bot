package repository

import (
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"time"
	"tradingViewWebhookBot/internal/domain"
)

type tradingStrategyRepository struct {
	db *sqlx.DB
}

func NewTradingStrategyRepository(db *sqlx.DB) TradingStrategy {
	return &tradingStrategyRepository{db: db}
}

func (r *tradingStrategyRepository) Create(strategy *domain.TradingStrategy) error {
	query := `
        INSERT INTO trading_strategies (name, description, created_at, updated_at)
        VALUES ($1, $2, $3, $3)
        RETURNING id`

	now := time.Now()
	return r.db.QueryRow(query, strategy.Name, strategy.Description, now).Scan(&strategy.Id)
}

func (r *tradingStrategyRepository) GetByID(id int64) (*domain.TradingStrategy, error) {
	strategy := &domain.TradingStrategy{}
	query := `
        SELECT id, name, description, created_at, updated_at
        FROM trading_strategies
        WHERE id = $1`

	err := r.db.QueryRow(query, id).Scan(
		&strategy.Id,
		&strategy.Name,
		&strategy.Description,
		&strategy.CreatedAt,
		&strategy.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return strategy, nil
}

func (r *tradingStrategyRepository) Update(strategy *domain.TradingStrategy) error {
	query := `
        UPDATE trading_strategies
        SET name = $1, description = $2, updated_at = $3
        WHERE id = $4`

	result, err := r.db.Exec(query,
		strategy.Name,
		strategy.Description,
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

func (r *tradingStrategyRepository) Delete(id int64) error {
	query := `DELETE FROM trading_strategies WHERE id = $1`
	result, err := r.db.Exec(query, id)
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
        SELECT id, name, description, created_at, updated_at
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
	query := `SELECT id, name, description, tag, enabled, created_at, updated_at 
              FROM trading_strategies 
              WHERE tag = $1 AND enabled = true`

	err := r.db.Get(&strategy, query, tag)
	if err != nil {
		if err == sql.ErrNoRows {
			zap.L().Warn("No strategy found", zap.String("tag", tag))
			return nil, nil // Return nil if no strategy found
		}
		return nil, fmt.Errorf("error finding trading strategy by tag: %w", err)
	}

	return &strategy, nil
}
