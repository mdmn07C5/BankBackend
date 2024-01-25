package postgresdb

import (
	"context"
	"fmt"
)

func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.connPool.Begin(ctx)
	if err != nil {
		return err
	}
	q := New(tx)
	err = fn(q)
	if err != nil {
		// rollback returns its own errors, combine and report both errors
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("Transaction error: %v, Rollback error: %v", err, rbErr)
		}
		return err // return original err if rollback is successful
	}

	return tx.Commit(ctx)
}
