package batcher

import (
	"btchrr/models"
	"errors"
)

// Batcher - encapsulates the batch size for splitting a slice into batches
type Batcher struct {
	batchSize    int
	placeholders map[string]struct{}
}

// NewBatcher - Creates a new batcher instance with the specified batch size.
// Returns an error if the batch size is invalid.
func NewBatcher(batchSize int) (*Batcher, error) {
	if batchSize <= 0 {
		return nil, models.ErrInvalidBatchSize
	}
	b := &Batcher{
		batchSize:    batchSize,
		placeholders: make(map[string]struct{}),
	}
	for _, ph := range []string{"?", "$", ":", ":"} {
		b.placeholders[ph] = struct{}{}
	}
	return b, nil
}

// BuildBatch - Builds a batch query from a single query
func (b *Batcher) BuildBatches(singleQuery string, items []any) ([]models.BatchedQuery, error) {
	batchedItems, err := b.batchItems(items)
	if err != nil {
		return []models.BatchedQuery{}, err
	}

	batches, err := b.BuildBatchQuery(singleQuery, b.batchSize, batchedItems)
	if err != nil {
		return []models.BatchedQuery{}, err
	}

	return batches, nil
}

// Batch - Splits the input slice into batches of the specified size.
// Returns an error if the input slice is empty.
func (b *Batcher) batchItems(items []any) ([][]any, error) {
	if len(items) == 0 {
		return nil, models.ErrNoItems
	}

	totalItems := len(items)
	numOfBatches := (totalItems + b.batchSize - 1) / b.batchSize // round up

	batches := make([][]any, 0, numOfBatches)

	for i := 0; i < totalItems; i += b.batchSize {
		end := i + b.batchSize
		if end > totalItems {
			end = totalItems
		}
		// Copy elements to avoid possible side effects if the original slice is modified
		batch := make([]any, end-i)
		copy(batch, items[i:end])
		batches = append(batches, batch)
	}

	return batches, nil
}

// BuildBatchQuery - builds a batch query from a single query
func (b *Batcher) BuildBatchQuery(singleQuery string, batchSize int, batchedItems [][]any) (batchedQueries []models.BatchedQuery, err error) {
	placeholder, err := b.detectPlaceholders(singleQuery)
	if err != nil {
		return []models.BatchedQuery{}, err
	}
	
	switch placeholder {
	case "?":
		return b.buildSqliteQuery(singleQuery, batchSize, batchedItems)
	case "$":
		return b.buildPostgresQuery(singleQuery, batchSize, batchedItems)
	default:
		return b.buildQueryWithNamedPlaceholders(singleQuery, batchSize, placeholder, batchedItems)
	}
}

// detectPlaceholder определяет тип плейсхолдера в SQL-запросе, ищет плейсхолдеры окружённые пробелами (" ? ", " $ ", " : ")
func (b *Batcher) detectPlaceholders(query string) (string, error) {
	for _, s := range query {
		if _, ok := b.placeholders[string(s)]; ok {
			return string(s), nil
		}
	}
	return "", models.ErrCannotDetectPlaceholder
}

func (b *Batcher) buildSqliteQuery(singleQuery string, batchSize int, batchedItems [][]any) (batchedQueries []models.BatchedQuery, err error) {
	return []models.BatchedQuery{}, errors.New("not implemented")
}

func (b *Batcher) buildPostgresQuery(singleQuery string, batchSize int, batchedItems [][]any) (batchedQueries []models.BatchedQuery, err error) {
	return []models.BatchedQuery{}, errors.New("not implemented")
}

func (b *Batcher) buildQueryWithNamedPlaceholders(singleQuery string, batchSize int, placeholder string, batchedItems [][]any) (batchedQueries []models.BatchedQuery, err error) {
	return []models.BatchedQuery{}, errors.New("not implemented")
}
