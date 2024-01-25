package postgresdb

import (
	"context"
	"testing"
	"time"

	"github.com/mdmn07C5/bank/util"
	"github.com/stretchr/testify/require"
)

func createRandomTransfer(t *testing.T, fromAccount, toAccount Account) Transfer {
	arg := CreateTransferParams{
		FromAccountID: fromAccount.ID,
		ToAccountID:   toAccount.ID,
		Amount:        util.RandomAmount(),
	}

	xfer, err := testStore.CreateTransfer(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, xfer)
	// sanity check
	require.Equal(t, arg.FromAccountID, xfer.FromAccountID)
	require.Equal(t, arg.ToAccountID, xfer.ToAccountID)
	require.Equal(t, arg.Amount, xfer.Amount)
	// postgres gen
	require.NotEmpty(t, xfer.ID)
	require.NotEmpty(t, xfer.CreatedAt)

	return xfer
}

func TestCreateTransfer(t *testing.T) {
	fromAccount := createRandomAccount(t)
	toAccount := createRandomAccount(t)
	createRandomTransfer(t, fromAccount, toAccount)
}

func TestGetTransfer(t *testing.T) {
	fromAccount := createRandomAccount(t)
	toAccount := createRandomAccount(t)
	xfer1 := createRandomTransfer(t, fromAccount, toAccount)

	xfer2, err := testStore.GetTransfer(context.Background(), xfer1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, xfer2)

	require.Equal(t, xfer1.ID, xfer2.ID)
	require.Equal(t, xfer1.FromAccountID, xfer2.FromAccountID)
	require.Equal(t, xfer1.ToAccountID, xfer2.ToAccountID)
	require.Equal(t, xfer1.Amount, xfer2.Amount)
	require.WithinDuration(t, xfer1.CreatedAt, xfer2.CreatedAt, time.Second)
}

func TestListTransfers(t *testing.T) {
	fromAccount := createRandomAccount(t)
	toAccount := createRandomAccount(t)

	for i := 0; i < 10; i++ {
		createRandomTransfer(t, fromAccount, toAccount)
		createRandomTransfer(t, toAccount, fromAccount)
	}

	arg := ListTransfersParams{
		FromAccountID: fromAccount.ID,
		ToAccountID:   fromAccount.ID,
		Limit:         5,
		Offset:        5,
	}

	xfers, err := testStore.ListTransfers(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, xfers, 5)

	for _, xfer := range xfers {
		require.NotEmpty(t, xfer)
		require.True(t, xfer.FromAccountID == fromAccount.ID || xfer.ToAccountID == fromAccount.ID)
	}
}
