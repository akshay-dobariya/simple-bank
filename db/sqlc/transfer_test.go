package db

import (
	"context"
	"testing"
	"time"

	"github.com/akshay-dobariya/simple-bank/util"
	"github.com/stretchr/testify/require"
)

func createRandomTransfer(t *testing.T, fromAccID, toAccID int64) Transfer {
	arg := CreateTransferParams{
		FromAccountID: fromAccID,
		ToAccountID:   toAccID,
		Amount:        util.RandomMoney(),
	}
	transfer, err := testQueries.CreateTransfer(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, transfer)

	require.Equal(t, transfer.FromAccountID, arg.FromAccountID)
	require.Equal(t, transfer.ToAccountID, arg.ToAccountID)
	require.Equal(t, transfer.Amount, arg.Amount)

	require.NotZero(t, transfer.ID)
	require.NotZero(t, transfer.CreatedAt)
	return transfer
}

func TestCreateTransfer(t *testing.T) {
	fromAccount := createRandomAccount(t)
	toAccount := createRandomAccount(t)
	createRandomTransfer(t, fromAccount.ID, toAccount.ID)

	// negative test case 1 - from acc id does not exists
	arg := CreateTransferParams{
		FromAccountID: util.RandomInt(1, 1000),
		ToAccountID:   toAccount.ID,
		Amount:        util.RandomMoney(),
	}
	transfer, err := testQueries.CreateTransfer(context.Background(), arg)
	require.Error(t, err)
	require.Empty(t, transfer)

	// negative test case 2 - to acc id does not exists
	arg = CreateTransferParams{
		FromAccountID: fromAccount.ID,
		ToAccountID:   util.RandomInt(1, 1000),
		Amount:        util.RandomMoney(),
	}
	transfer, err = testQueries.CreateTransfer(context.Background(), arg)
	require.Error(t, err)
	require.Empty(t, transfer)
}

func TestGetTransfer(t *testing.T) {
	fromAccount := createRandomAccount(t)
	toAccount := createRandomAccount(t)
	transfer1 := createRandomTransfer(t, fromAccount.ID, toAccount.ID)
	transfer2, err := testQueries.GetTransfer(context.Background(), transfer1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, transfer2)

	require.Equal(t, transfer1.ID, transfer2.ID)
	require.Equal(t, transfer1.Amount, transfer2.Amount)
	require.Equal(t, transfer1.FromAccountID, transfer2.FromAccountID)
	require.Equal(t, transfer1.ToAccountID, transfer2.ToAccountID)
	require.WithinDuration(t, transfer1.CreatedAt, transfer2.CreatedAt, time.Second)
}

func TestListTransfers(t *testing.T) {
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			createRandomTransfer(t, account1.ID, account2.ID)
		} else {
			createRandomTransfer(t, account2.ID, account1.ID)
		}
	}
	arg := ListTransfersParams{
		FromAccountID: account1.ID,
		ToAccountID:   account1.ID,
		Limit:         5,
		Offset:        5,
	}
	transfers, err := testQueries.ListTransfers(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, transfers, 5)
	for _, entry := range transfers {
		require.NotEmpty(t, entry)
	}

	// negative testcase - invalid acc id
	arg = ListTransfersParams{
		FromAccountID: util.RandomInt(1, 1000),
		ToAccountID:   util.RandomInt(1, 1000),
		Limit:         5,
		Offset:        5,
	}
	transfers, err = testQueries.ListTransfers(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, transfers, 0)
}
