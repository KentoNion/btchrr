package dbadapter

import (
	"context"
	"database/sql"

	"btchrr/models"
)

type DBAdapter struct {
	db *sql.DB
}

func NewSQLAdapter(db *sql.DB) *DBAdapter {
	return &DBAdapter{db: db}
}

func (a *DBAdapter) Exec(ctx context.Context, batchedQuery models.BatchedQuery) (sql.Result, error) {
	res, err := a.db.ExecContext(ctx, string(batchedQuery))

	return res, err
}

// CheckQuery - checks if the query is valid
func (a *DBAdapter) CheckQuery(query string) error {
	stmt, err := a.db.Prepare(query) 
	if err != nil {
    return err
	}
	stmt.Close()
	return nil
}
