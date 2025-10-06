package Btchrr

import (
	"btchrr/batcher"
	"btchrr/dbadapter"
	"context"
	"database/sql"

	"btchrr/models"
)

// Btchrr - main struct for the package
type Btchrr struct {
	batcher     *batcher.Batcher
	executor    Executor
}

// Executor - interface for the sql execution
type Executor interface {
	Exec(ctx context.Context, query models.BatchedQuery) (sql.Result, error)
	CheckQuery(query string) error
}

// AggregatedResult - aggregated result from all batches
type AggregatedResult struct {
	rowsAffected int64
	lastInsertId int64
}

func (r *AggregatedResult) LastInsertId() (int64, error) {
	return r.lastInsertId, nil
}

func (r *AggregatedResult) RowsAffected() (int64, error) {
	return r.rowsAffected, nil
}

// NewBtchrr - creates a new Btchrr instance
func NewBtchrr(batchSize int, db *sql.DB) (*Btchrr, error) {
	batcher, err := batcher.NewBatcher(batchSize)
	if err != nil {
		return nil, err
	}

	dbAdapter := dbadapter.NewSQLAdapter(db)

	return &Btchrr{
		batcher:     batcher,
		executor:    dbAdapter,
	}, nil
}

// Exec - accepts a query for single item, items and executes it in batches
func (b *Btchrr) Exec(ctx context.Context, query string, items []any) (sql.Result, error) {
	err := b.executor.CheckQuery(query)
	if err != nil {
		return nil, err
	}

	batches, err := b.batcher.BuildBatches(query, items)
	if err != nil {
		return nil, err
	}

	var totalRowsAffected int64
	var lastInsertId int64

	// Execute SQL query for each batch (transforming single-item query to batch query)
	for _, batch := range batches {

		result, err := b.executor.Exec(ctx, batch)
		if err != nil {
			return nil, err
		}

		// Суммируем результаты от всех батчей
		rowsAffected, _ := result.RowsAffected()
		totalRowsAffected += rowsAffected

		// Берем последний InsertId
		insertId, _ := result.LastInsertId()
		if insertId > 0 {
			lastInsertId = insertId
		}
	}

	return &AggregatedResult{
		rowsAffected: totalRowsAffected,
		lastInsertId: lastInsertId,
	}, nil
}
