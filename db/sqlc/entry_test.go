package db

import (
	"context"
	"testing"
	"time"

	"github.com/akshay-dobariya/simple-bank/util"
	"github.com/stretchr/testify/require"
)

func createRandomEntry(t *testing.T, accID int64) Entry {
	arg := CreateEntryParams{
		AccountID: accID,
		Amount:    util.RandomMoney(),
	}
	entry, err := testQueries.CreateEntry(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, entry)

	require.Equal(t, entry.AccountID, arg.AccountID)
	require.Equal(t, entry.Amount, arg.Amount)

	require.NotZero(t, entry.ID)
	require.NotZero(t, entry.CreatedAt)
	return entry
}

func TestCreateEntry(t *testing.T) {
	account := createRandomAccount(t)
	createRandomEntry(t, account.ID)

	// negative test case - acc id does not exists
	arg := CreateEntryParams{
		AccountID: util.RandomInt(1, 1000),
		Amount:    util.RandomMoney(),
	}
	entry, err := testQueries.CreateEntry(context.Background(), arg)
	require.Error(t, err)
	require.Empty(t, entry)
}

func TestGetEntry(t *testing.T) {
	account := createRandomAccount(t)
	entry1 := createRandomEntry(t, account.ID)
	entry2, err := testQueries.GetEntry(context.Background(), entry1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, entry2)

	require.Equal(t, entry1.ID, entry2.ID)
	require.Equal(t, entry1.Amount, entry2.Amount)
	require.Equal(t, entry1.AccountID, entry2.AccountID)
	require.WithinDuration(t, entry1.CreatedAt, entry2.CreatedAt, time.Second)
}

func TestListEntries(t *testing.T) {
	account := createRandomAccount(t)
	for i := 0; i < 10; i++ {
		createRandomEntry(t, account.ID)
	}
	arg := ListEntriesParams{
		AccountID: account.ID,
		Limit:     5,
		Offset:    5,
	}
	entries, err := testQueries.ListEntries(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, entries, 5)
	for _, entry := range entries {
		require.NotEmpty(t, entry)
	}

	// negative testcase - invalid acc id
	arg = ListEntriesParams{
		AccountID: util.RandomInt(1, 1000),
		Limit:     5,
		Offset:    5,
	}
	entries, err = testQueries.ListEntries(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, entries, 0)
}
