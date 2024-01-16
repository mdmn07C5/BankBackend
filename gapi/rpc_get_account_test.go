package gapi

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	mockdb "github.com/mdmn07C5/bank/db/mock"
	db "github.com/mdmn07C5/bank/db/sqlc"
	"github.com/mdmn07C5/bank/token"

	"github.com/mdmn07C5/bank/pb"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestGetAccountAPI(t *testing.T) {
	user, _ := randomUser(t)
	account := randomAccount(user.Username)

	testCases := []struct {
		name          string
		req           *pb.GetAccountRequest
		buildStubs    func(store *mockdb.MockStore)
		buildContext  func(t *testing.T, tokenMaker token.Maker) context.Context
		checkResponse func(t *testing.T, res *pb.GetAccountResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.GetAccountRequest{
				Id: account.ID,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				accessToken, _, err := tokenMaker.CreateToken(user.Username, time.Minute)
				require.NoError(t, err)
				bearerToken := fmt.Sprintf("%s %s", authorizationBearer, accessToken)
				md := metadata.MD{
					authorizationHeader: []string{
						bearerToken,
					},
				}
				return metadata.NewIncomingContext(context.Background(), md)
			},
			checkResponse: func(t *testing.T, res *pb.GetAccountResponse, err error) {
				require.NoError(t, err)
				require.Equal(t, account.ID, res.Account.Id)
				require.Equal(t, account.Owner, res.Account.Owner)
				require.NotNil(t, res)
			},
		},
		{
			name: "NoAuthorization",
			req: &pb.GetAccountRequest{
				Id: account.ID,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return context.Background()
			},
			checkResponse: func(t *testing.T, res *pb.GetAccountResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "UnauthorizedUser",
			req: &pb.GetAccountRequest{
				Id: account.ID,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				unauthorizedUser, _ := randomUser(t)
				accessToken, _, err := tokenMaker.CreateToken(unauthorizedUser.Username, time.Minute)
				require.NoError(t, err)
				bearerToken := fmt.Sprintf("%s %s", authorizationBearer, accessToken)
				md := metadata.MD{
					authorizationHeader: []string{
						bearerToken,
					},
				}
				return metadata.NewIncomingContext(context.Background(), md)
			},
			checkResponse: func(t *testing.T, res *pb.GetAccountResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.PermissionDenied, st.Code())
			},
		},
		{
			name: "InternalError",
			req: &pb.GetAccountRequest{
				Id: account.ID,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(account, sql.ErrConnDone)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				accessToken, _, err := tokenMaker.CreateToken(user.Username, time.Minute)
				require.NoError(t, err)
				bearerToken := fmt.Sprintf("%s %s", authorizationBearer, accessToken)
				md := metadata.MD{
					authorizationHeader: []string{
						bearerToken,
					},
				}
				return metadata.NewIncomingContext(context.Background(), md)
			},
			checkResponse: func(t *testing.T, res *pb.GetAccountResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, st.Code())
			},
		},
		{
			name: "AccountNotFound",
			req: &pb.GetAccountRequest{
				Id: account.ID,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				accessToken, _, err := tokenMaker.CreateToken(user.Username, time.Minute)
				require.NoError(t, err)
				bearerToken := fmt.Sprintf("%s %s", authorizationBearer, accessToken)
				md := metadata.MD{
					authorizationHeader: []string{
						bearerToken,
					},
				}
				return metadata.NewIncomingContext(context.Background(), md)
			},
			checkResponse: func(t *testing.T, res *pb.GetAccountResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.NotFound, st.Code())
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

			ctx := tc.buildContext(t, server.tokenMaker)
			res, err := server.GetAccount(ctx, tc.req)
			tc.checkResponse(t, res, err)
		})
	}
}
