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

func TestListAccountsAPI(t *testing.T) {
	n := int32(10)
	user, _ := randomUser(t)
	accounts := make([]db.Account, n)
	for i := int32(0); i < n; i++ {
		accounts[i] = randomAccount(user.Username)

	}
	start := int32(1)

	testCases := []struct {
		name          string
		req           *pb.ListAccountsRequest
		buildStubs    func(store *mockdb.MockStore)
		buildContext  func(t *testing.T, tokenMaker token.Maker) context.Context
		checkResponse func(t *testing.T, res *pb.ListAccountsResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.ListAccountsRequest{
				PageId:   start,
				PageSize: n,
			},
			buildStubs: func(store *mockdb.MockStore) {
				listAccountParams := db.ListAccountsParams{
					Owner:  user.Username,
					Limit:  n,
					Offset: start - 1,
				}
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(listAccountParams)).
					Times(1).
					Return(accounts, nil)
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
			checkResponse: func(t *testing.T, res *pb.ListAccountsResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				accs := res.GetAccounts()
				require.Len(t, accs, int(n))
				requireResponseAccountsMatch(t, accounts, accs)
			},
		},
		{
			name: "NoAuthorization",
			req: &pb.ListAccountsRequest{
				PageId:   start,
				PageSize: n,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return context.Background()
			},
			checkResponse: func(t *testing.T, res *pb.ListAccountsResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "InternalError",
			req: &pb.ListAccountsRequest{
				PageId:   start,
				PageSize: n,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(1).
					Return(accounts, sql.ErrConnDone)
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
			checkResponse: func(t *testing.T, res *pb.ListAccountsResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, st.Code())
			},
		},
		{
			name: "InvalidPageID",
			req: &pb.ListAccountsRequest{
				PageId:   -1,
				PageSize: n,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
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
			checkResponse: func(t *testing.T, res *pb.ListAccountsResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		// TODO: Test case for unauthorized account
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
			res, err := server.ListAccounts(ctx, tc.req)
			tc.checkResponse(t, res, err)
		})
	}
}

func requireResponseAccountsMatch(t *testing.T, accounts []db.Account, response []*pb.Account) {
	require.Equal(t, len(accounts), len(response))
	accs := map[int64]bool{}
	for i := range accounts {
		accs[accounts[i].ID] = true
	}
	for i := range response {
		require.True(t, accs[response[i].Id])
	}
}
