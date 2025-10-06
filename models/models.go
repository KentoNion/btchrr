package models

import "errors"

var (
	ErrNoItems                 = errors.New("no items recieved, batch is empty")
	ErrInvalidBatchSize        = errors.New("batch size must be greater than zero")
	ErrCannotDetectPlaceholder = errors.New("cannot detect placeholder in query")
)

// BatchedQuery - single batch query
type BatchedQuery string
