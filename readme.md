# Btchrr

A Go package for automatic batching of database operations with query transformation and result aggregation.

## Features

- ✅ Automatic splitting of items into batches of specified size
- ✅ SQL query transformation from single-item to batch queries
- ✅ Batch query execution with result aggregation
- ✅ Support for any SQL database through `Executor` interface
- ✅ Database-specific placeholder support (PostgreSQL, MySQL, SQLite, etc.)
- ✅ Error handling and input validation

## Usage

### PostgreSQL Example

```go
package main

import (
    "context"
    "database/sql"
    "log"
    "your-project/btchrr"  // Path to your package in the project
)

func main() {
    // Database connection
    db, err := sql.Open("postgres", "connection_string")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Create Btchrr instance with batch size 100 and PostgreSQL placeholder
    btchrr, err := btchrr.NewBtchrr(100, db, "$1")
    if err != nil {
        log.Fatal(err)
    }

    // Prepare data for insertion
    items := []any{"John", "Jane", "Bob", "Alice", "Charlie"}

    // Execute query with automatic batching
    ctx := context.Background()
    result, err := btchrr.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", items)
    if err != nil {
        log.Fatal(err)
    }

    // Get aggregated results
    rowsAffected, _ := result.RowsAffected()
    lastId, _ := result.LastInsertId()

    log.Printf("Rows affected: %d, Last ID: %d", rowsAffected, lastId)
}
```

### MySQL/SQLite Example

```go
// For MySQL or SQLite, use "?" placeholder
btchrr, err := btchrr.NewBtchrr(100, db, "?")

// Query will be transformed from:
// "INSERT INTO users (name) VALUES (?)"
// To:
// "INSERT INTO users (name) VALUES (?, ?, ?)"
```

## API

### NewBtchrr(batchSize int, db *sql.DB, placeholder string) (*Btchrr, error)

Creates a new Btchrr instance with specified batch size, database connection, and placeholder format.

**Parameters:**

- `batchSize` - batch size (must be > 0)
- `db` - database connection
- `placeholder` - database-specific placeholder ("?" for MySQL/SQLite, "$1" for PostgreSQL)

**Returns:**

- `*Btchrr` - Btchrr instance
- `error` - creation error

### Exec(ctx context.Context, query string, items []any) (sql.Result, error)

Executes SQL query for each batch and returns aggregated result.

**Parameters:**

- `ctx` - execution context
- `query` - SQL query for single item
- `items` - slice of items to process

**Returns:**

- `sql.Result` - aggregated result from all batches
- `error` - execution error

## Future Plans

- [ ] Support for GORM, sqlx, ent, go-pg, pgx
- [ ] Dynamic batch size based on item count
- [ ] Transaction support
- [ ] Performance metrics
- [ ] Transaction support
