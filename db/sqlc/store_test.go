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

	exists := make(map[int]bool)

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
		// check accounts
		fromAccount := res.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, fromAccount.ID, res.FromAccount.ID)

		toAccount := res.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, toAccount.ID, res.ToAccount.ID)

		// check accounts balance
		delta1 := account1.Balance - fromAccount.Balance
		delta2 := toAccount.Balance - account2.Balance
		require.Equal(t, delta1, delta2)
		require.True(t, delta1 > 0)
		require.True(t, delta1%amount == 0)

		k := int(delta1 / amount)
		require.True(t, 1 <= k && k <= n)
		require.NotContains(t, exists, k)
		exists[k] = true
	}

	// check final balances
	updatedAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	require.Equal(t, account1.Balance-int64(n)*amount, updatedAccount1.Balance)
	require.Equal(t, account2.Balance+int64(n)*amount, updatedAccount2.Balance)

}

func TestTransferTxDeadlock(t *testing.T) {
	store := NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	n := 10
	amount := int64(10)

	// see below
	errs := make(chan error)

	for i := 0; i < n; i++ {
		fromAccountID := account1.ID
		toAccountID := account2.ID

		if i%2 == 1 {
			fromAccountID, toAccountID = account2.ID, account1.ID
		}

		go func() {
			_, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: fromAccountID,
				ToAccountID:   toAccountID,
				Amount:        amount,
			})
			errs <- err
		}()
	}

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
	}

	// check final balances
	updatedAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	require.Equal(t, account1.Balance, updatedAccount1.Balance)
	require.Equal(t, account2.Balance, updatedAccount2.Balance)

}
