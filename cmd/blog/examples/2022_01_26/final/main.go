package main

import (
	"context"
	"database/sql"
	"errors"

	_ "github.com/go-sql-driver/mysql"
)

// SQLTokenRequester implements the TokenRequester interface
// by using a sql compliant database as persistence layer.
type SQLTokenRequester struct {
	db       *sql.DB
	bucketID string
}

// NewSQLTokenRequester creates a new SQLTokenRequester and
// returns a pointer to it.
// Considering that the bucket is already created.
func NewSQLTokenRequester(ctx context.Context, db *sql.DB, bucketID string) *SQLTokenRequester {
	return &SQLTokenRequester{
		db:       db,
		bucketID: bucketID,
	}
}

// RequestToken requests a token from the shared bucket.
// This implementation assumes that the token bucket and the stored
// procedure were created beforehand in the database.
func (tr *SQLTokenRequester) RequestToken(ctx context.Context) error {
	q := `
		CALL request_token(?);
	`

	result, err := tr.db.ExecContext(ctx, q, tr.bucketID)
	if err != nil {
		return err
	}

	// If a row is affected, the stored procedure call was successful,
	// meaning that a token was given. If an error occurs or no
	// rows were affected the procedure last update was not successful
	// therefore no token was available.
	if rows, err := result.RowsAffected(); err != nil || rows <= 0 {
		return errors.New("no token available")
	}

	return nil
}
