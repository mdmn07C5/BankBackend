package postgresdb

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	n := 10
	amount := int64(10)

	// see below
	errs := make(chan error)
	results := make(chan TransferTxResult)

	for i := 0; i < n; i++ {
		go func() {
			// you cannot use testify to check here because this function is running
			// from a different goroutine form the one that the calling (TestTransferTX)
			// is running on so there's no guarantee that it will stop the whole test
			// if a condition is not satisfied
			result, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})
			// instead, we'll send them back to the main goroutine that our test is running on
			// and check them from there
			errs <- err
			results <- result
		}()
	}

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		res := <-results
		require.NotEmpty(t, res)

		// check transfer
		xfer := res.Transfer
		require.NotEmpty(t, xfer)
		require.Equal(t, account1.ID, xfer.FromAccountID)
		require.Equal(t, account2.ID, xfer.ToAccountID)
		require.Equal(t, xfer.Amount, amount)
		require.NotZero(t, xfer.ID)
		require.NotZero(t, xfer.CreatedAt)

		//check if it really exists in the db
		_, err = store.GetTransfer(context.Background(), xfer.ID)
		require.NoError(t, err)

		// check from entry
		fromEntry := res.FromEntry
		require.NotEmpty(t, xfer)
		require.Equal(t, account1.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		//check if entry really exists in db
		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		// check to entry
		toEntry := res.ToEntry
		require.NotEmpty(t, xfer)
		require.Equal(t, account2.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		//check if entry really exists in db
		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		// TODO: check account's balance
	}

}
