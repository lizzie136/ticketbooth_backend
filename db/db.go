package db

import (
	"github.com/jmoiron/sqlx"
)

// DB wraps sqlx.DB to provide database operations
type DB struct {
	*sqlx.DB
}

// New creates a new DB instance
func New(sqlxDB *sqlx.DB) *DB {
	return &DB{DB: sqlxDB}
}

// WithTx executes a function within a transaction
func (db *DB) WithTx(fn func(*sqlx.Tx) error) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(tx)
	return err
}

