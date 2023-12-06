package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	mockdb "github.com/mdmn07C5/bank/db/mock"
	db "github.com/mdmn07C5/bank/db/sqlc"
	"github.com/mdmn07C5/bank/util"
	"github.com/stretchr/testify/require"
)

func TestTransferAPI(t *testing.T) {
	fromAccount := randomAccount()
	toAccount := randomAccount()
	otherAccount := randomAccount()

	fromAccount.Currency = util.USD
	toAccount.Currency = util.USD
	otherAccount.Currency = util.MXN

	amount := int64(10)

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(mockStore *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			buildStubs: func(mockStore *mockdb.MockStore) {
				mockStore.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(fromAccount, nil)
				mockStore.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).
					Times(1).
					Return(toAccount, nil)

				tx := db.TransferTxParams{
					FromAccountID: fromAccount.ID,
					ToAccountID:   toAccount.ID,
					Amount:        amount,
				}
				mockStore.EXPECT().
					TransferTx(gomock.Any(), gomock.Eq(tx)).
					Times(1)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "FromAccountNotFound",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			buildStubs: func(mockStore *mockdb.MockStore) {
				mockStore.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(fromAccount, nil)
				mockStore.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)

			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "ToAccountNotFound",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			buildStubs: func(mockStore *mockdb.MockStore) {
				mockStore.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)

			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "UnsupportedCurrency",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          amount,
				"currency":        "JPY",
			},
			buildStubs: func(mockStore *mockdb.MockStore) {
				mockStore.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "ToAccountCurrencyMismatch",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   otherAccount.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			buildStubs: func(mockStore *mockdb.MockStore) {
				mockStore.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(fromAccount, nil)
				mockStore.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(otherAccount.ID)).
					Times(1).
					Return(otherAccount, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "FromAccountCurrencyMismatch",
			body: gin.H{
				"from_account_id": otherAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			buildStubs: func(mockStore *mockdb.MockStore) {
				mockStore.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(otherAccount.ID)).
					Times(1).
					Return(otherAccount, nil)
				mockStore.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "TransferTxError",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			buildStubs: func(mockStore *mockdb.MockStore) {
				mockStore.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(fromAccount, nil)
				mockStore.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).
					Times(1).
					Return(toAccount, nil)

				mockStore.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.TransferTxResult{}, sql.ErrTxDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "GetAccountError",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			buildStubs: func(mockStore *mockdb.MockStore) {
				mockStore.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockStore := mockdb.NewMockStore(ctrl)

			tc.buildStubs(mockStore)

			server := newTestServer(t, mockStore)
			recorder := httptest.NewRecorder()

			body, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/transfers"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}

}
