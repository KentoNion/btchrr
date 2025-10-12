package batcher

import (
	"strings"
	"errors"
	"strconv"

	"btchrr/models"
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
	for _, ph := range []string{"?", "$", ":"} {
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

	batches, err := b.BuildBatchQuery(singleQuery, b.batchSize, batchedItems, len(batchedItems))
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
func (b *Batcher) BuildBatchQuery(singleQuery string, batchSize int, batchedItems [][]any, numOfBatches int) (batchedQueries []models.BatchedQuery, err error) {
	placeholder, err := b.detectPlaceholders(singleQuery)
	if err != nil {
		return []models.BatchedQuery{}, err
	}

	switch placeholder {
	case "?":
		for i :=0; i < numOfBatches; i++ {
			batchQuery, err := b.buildSqliteQuery(singleQuery, batchedItems[i])
			if err != nil {
				return []models.BatchedQuery{}, err
			}
			batchedQueries = append(batchedQueries, batchQuery)
		}
		return
	case "$":
		for i :=0; i < numOfBatches; i++ {
			batchQuery, err := b.buildPostgresQuery(singleQuery, batchedItems[i])
			if err != nil {
				return []models.BatchedQuery{}, err
			}
			batchedQueries = append(batchedQueries, batchQuery)
		}
		return
	default:
		return b.buildQueryWithNamedPlaceholders(singleQuery, placeholder, batchedItems)
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

func (b *Batcher) buildSqliteQuery(singleQuery string, batchedItems []any) (batchedQueries models.BatchedQuery, err error) {
	// INSERT INTO users VALUES (?, ?);
	idx := strings.Index(strings.ToLower(singleQuery), "values")
	if idx == -1 {
		return "", errors.New("no values found in query")
	}

	prefix := singleQuery[:idx+6] // +6 - values
	valuesPart := singleQuery[idx+6:]

	numOfInserts := strings.Count(valuesPart, "?")
	if numOfInserts == 0 {
		return "", errors.New("no placeholders found in query")
	}

	singleItemValues := "(" + strings.Repeat("?, ", numOfInserts - 1) + "?)"
	//INSERT INTO users (name, age) VALUES (?, ?), (?, ?), (?, ?);
	newValues := strings.Repeat(singleItemValues + ", ", b.batchSize - 1) + singleItemValues
	newQuery := prefix + " " + newValues + ";"


	return models.BatchedQuery(newQuery), nil
}

func (b *Batcher) buildPostgresQuery(singleQuery string, batchedItems []any) (batchedQueries models.BatchedQuery, err error) {
	// INSERT INTO users VALUES ($1, $2);
	idx := strings.Index(strings.ToLower(singleQuery), "values")
	if idx == -1 {
		return "", errors.New("no values found in query")
	}

	prefix := singleQuery[:idx+6] // +6 - values
	valuesPart := singleQuery[idx+6:]

	numOfInserts := strings.Count(valuesPart, "$")
	if numOfInserts == 0 {
		return "", errors.New("no placeholders found in query")
	}

	//INSERT INTO users VALUES ($1, $2), ($3, $4), ($5, $6);
	placeholderNum := 1
	query := ""
	for i := 0; i < b.batchSize ; i++ {
		if i > 0 {
			query += ", "
		}
		query += "("
		for j := 0; j < numOfInserts; j++ {
			query += "$" + strconv.Itoa(placeholderNum) + ", "
			placeholderNum++
		}
		query += ")"
	}

	query = prefix + " " + query + ";"

	return models.BatchedQuery(query), nil
}

func (b *Batcher) buildQueryWithNamedPlaceholders(singleQuery string, placeholder string, batchedItems [][]any) (batchedQueries []models.BatchedQuery, err error) {
	return []models.BatchedQuery{}, errors.New("not implemented")
}
