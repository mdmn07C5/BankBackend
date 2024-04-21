package gapi

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	mockdb "github.com/mdmn07C5/bank/db/mock"
	db "github.com/mdmn07C5/bank/db/sqlc"
	"github.com/mdmn07C5/bank/token"
	"github.com/mdmn07C5/bank/util"
	mockwk "github.com/mdmn07C5/bank/worker/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/mdmn07C5/bank/pb"
	"github.com/stretchr/testify/require"
)

func TestTransferFundsAPI(t *testing.T) {
	fromUser, _ := randomUser(t)
	toUser, _ := randomUser(t)
	otherUser, _ := randomUser(t)

	fromAccount := randomAccount(fromUser.Username)
	toAccount := randomAccount(toUser.Username)
	otherAccount := randomAccount(otherUser.Username)

	fromAccount.Currency = util.USD
	toAccount.Currency = util.USD
	otherAccount.Currency = util.MXN

	amount := int64(10)

	testCases := []struct {
		name          string
		req           *pb.TransferRequest
		buildStubs    func(mockStore *mockdb.MockStore)
		buildContext  func(t *testing.T, tokenMaker token.Maker) context.Context
		checkResponse func(t *testing.T, res *pb.TransferResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.TransferRequest{
				FromAccountId: fromAccount.ID,
				ToAccountId:   toAccount.ID,
				Amount:        amount,
				Currency:      util.USD,
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
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, fromUser.Username, fromUser.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.TransferResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
			},
		},
		{
			name: "Unauthenticated",
			req: &pb.TransferRequest{
				FromAccountId: fromAccount.ID,
				ToAccountId:   toAccount.ID,
				Amount:        amount,
				Currency:      util.USD,
			},
			buildStubs: func(mockStore *mockdb.MockStore) {
				mockStore.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(fromAccount, nil)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, toUser.Username, toUser.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.TransferResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "FromAccountNotFound",
			req: &pb.TransferRequest{
				FromAccountId: fromAccount.ID,
				ToAccountId:   toAccount.ID,
				Amount:        amount,
				Currency:      util.USD,
			},
			buildStubs: func(mockStore *mockdb.MockStore) {
				mockStore.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(db.Account{}, db.ErrRecordNotFound)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, fromUser.Username, fromUser.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.TransferResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.NotFound, st.Code())
			},
		},
		{
			name: "ToAccountNotFound",
			req: &pb.TransferRequest{
				FromAccountId: fromAccount.ID,
				ToAccountId:   toAccount.ID,
				Amount:        amount,
				Currency:      util.USD,
			},
			buildStubs: func(mockStore *mockdb.MockStore) {
				mockStore.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(fromAccount, nil)
				mockStore.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).
					Times(1).
					Return(db.Account{}, db.ErrRecordNotFound)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, fromUser.Username, fromUser.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.TransferResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.NotFound, st.Code())
			},
		},
		{
			name: "CurrencyMismatch",
			req: &pb.TransferRequest{
				FromAccountId: fromAccount.ID,
				ToAccountId:   otherAccount.ID,
				Amount:        amount,
				Currency:      otherAccount.Currency,
			},
			buildStubs: func(mockStore *mockdb.MockStore) {
				mockStore.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(fromAccount, nil)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, fromUser.Username, fromUser.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.TransferResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "UnauthorizedUser",
			req: &pb.TransferRequest{
				FromAccountId: fromAccount.ID,
				ToAccountId:   toAccount.ID,
				Amount:        amount,
				Currency:      util.USD,
			},
			buildStubs: func(mockStore *mockdb.MockStore) {
				mockStore.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(fromAccount, nil)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, otherUser.Username, otherUser.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.TransferResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "InternalError",
			req: &pb.TransferRequest{
				FromAccountId: fromAccount.ID,
				ToAccountId:   toAccount.ID,
				Amount:        amount,
				Currency:      util.USD,
			},
			buildStubs: func(mockStore *mockdb.MockStore) {
				mockStore.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, fromUser.Username, fromUser.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.TransferResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, st.Code())
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockStore := mockdb.NewMockStore(ctrl)

			tskCtrl := gomock.NewController(t)
			defer tskCtrl.Finish()
			taskDistributor := mockwk.NewMockTaskDistributor(tskCtrl)

			tc.buildStubs(mockStore)
			server := newTestServer(t, mockStore, taskDistributor)

			ctx := tc.buildContext(t, server.tokenMaker)
			res, err := server.TransferFunds(ctx, tc.req)
			tc.checkResponse(t, res, err)
		})
	}
}
