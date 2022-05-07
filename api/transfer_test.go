package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	mockdb "github.com/akshay-dobariya/simple-bank/db/mock"
	db "github.com/akshay-dobariya/simple-bank/db/sqlc"
	"github.com/akshay-dobariya/simple-bank/util"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func randomAccountWithCurrency(currency string) db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    util.RandomOwner(),
		Balance:  util.RandomMoney(),
		Currency: currency,
	}
}

func randomEntry(accountID, amount int64) db.Entry {
	return db.Entry{
		ID:        util.RandomInt(1, 1000),
		AccountID: accountID,
		Amount:    amount,
	}
}

func TestCreateTransferAPI(t *testing.T) {
	transferJSONStr := `{
		"from_account_id": %d,
		"to_account_id": %d,
		"currency": "%s",
		"amount": %d
	}`
	amount := int64(20)
	currency := util.RandomCurrency()
	account1 := randomAccountWithCurrency(currency)
	account2 := randomAccountWithCurrency(currency)
	entry1 := randomEntry(account1.ID, -amount)
	entry2 := randomEntry(account2.ID, amount)
	result := db.TransferTxResult{
		Transfer: db.Transfer{
			ID:            util.RandomInt(1, 1000),
			FromAccountID: account1.ID,
			ToAccountID:   account2.ID,
			Amount:        amount,
		},
		FromAccount: account1,
		ToAccount:   account2,
		FromEntry:   entry1,
		ToEntry:     entry2,
	}

	testCases := []struct {
		name          string
		transferJSON  json.RawMessage
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			transferJSON: json.RawMessage(
				fmt.Sprintf(transferJSONStr, account1.ID, account2.ID, currency, amount),
			),
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					TransferTx(gomock.Any(), db.TransferTxParams{
						FromAccountID: account1.ID,
						ToAccountID:   account2.ID,
						Amount:        amount,
					}).
					Times(1).
					Return(result, nil).After(
					store.EXPECT().
						GetAccount(gomock.Any(), gomock.Eq(account2.ID)).
						Times(1).
						Return(account2, nil).
						After(
							store.EXPECT().
								GetAccount(gomock.Any(), gomock.Eq(account1.ID)).
								Times(1).
								Return(account1, nil),
						),
				)

			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchTxResult(t, recorder.Body, result)
			},
		},
		{
			name: "InvalidRequestBodyWithUnsupportedCurrency",
			transferJSON: json.RawMessage(
				fmt.Sprintf(transferJSONStr, account1.ID+1000, account2.ID, "ABC", amount),
			),
			buildStubs: func(store *mockdb.MockStore) {},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "FromAccountNotFound",
			transferJSON: json.RawMessage(
				fmt.Sprintf(transferJSONStr, account1.ID+1000, account2.ID, currency, amount),
			),
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account1.ID+1000)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "ToAccountInternalError",
			transferJSON: json.RawMessage(
				fmt.Sprintf(transferJSONStr, account1.ID, account2.ID+1000, currency, amount),
			),
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account2.ID+1000)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone).
					After(
						store.EXPECT().
							GetAccount(gomock.Any(), gomock.Eq(account1.ID)).
							Times(1).
							Return(account1, nil),
					)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "FromAccountInvalidCurrency",
			transferJSON: json.RawMessage(
				fmt.Sprintf(transferJSONStr, account1.ID, account2.ID, currency, amount),
			),
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account1.ID)).
					Times(1).
					Return(db.Account{Currency: getInvalidCurrency(currency)}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InternalErrorInTransferTx",
			transferJSON: json.RawMessage(
				fmt.Sprintf(transferJSONStr, account1.ID, account2.ID, currency, amount),
			),
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					TransferTx(gomock.Any(), db.TransferTxParams{
						FromAccountID: account1.ID,
						ToAccountID:   account2.ID,
						Amount:        amount,
					}).
					Times(1).
					Return(db.TransferTxResult{}, sql.ErrConnDone).After(
					store.EXPECT().
						GetAccount(gomock.Any(), gomock.Eq(account2.ID)).
						Times(1).
						Return(account2, nil).
						After(
							store.EXPECT().
								GetAccount(gomock.Any(), gomock.Eq(account1.ID)).
								Times(1).
								Return(account1, nil),
						),
				)

			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			// start test server and send request
			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := "/transfers"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(tc.transferJSON))
			require.NoError(t, err)
			request.Header.Set("Content-Type", "application/json")

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func requireBodyMatchTxResult(t *testing.T, body *bytes.Buffer, result db.TransferTxResult) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)
	t.Logf("Received response - '%s'", string(data))

	var gotTxResult db.TransferTxResult
	err = json.Unmarshal(data, &gotTxResult)
	require.NoError(t, err)
	require.Equal(t, gotTxResult, result)
}

func getInvalidCurrency(origCurrency string) string {
	switch origCurrency {
	case util.USD:
		return util.INR
	default:
		return util.USD
	}
}
