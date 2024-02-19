package gapi

import (
	"context"
	"database/sql"
	"testing"

	"github.com/golang/mock/gomock"
	mockdb "github.com/mdmn07C5/bank/db/mock"
	db "github.com/mdmn07C5/bank/db/sqlc"
	"github.com/mdmn07C5/bank/pb"
	mockwk "github.com/mdmn07C5/bank/worker/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestLoginUserAPI(t *testing.T) {
	user, password := randomVerifiedUser(t)
	unverifiedUser, unverifedUserPW := randomUser(t)

	testCases := []struct {
		name          string
		req           *pb.LoginUserRequest
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, res *pb.LoginUserResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.LoginUserRequest{
				Username: user.Username,
				Password: password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(user, nil)

				store.EXPECT().
					CreateSession(gomock.Any(), gomock.Any()).
					Times(1)
			},
			checkResponse: func(t *testing.T, res *pb.LoginUserResponse, err error) {
				require.NoError(t, err)
				require.Equal(t, user.Username, res.User.Username)
				require.NotNil(t, res)
				require.NotNil(t, res.GetAccessToken())
				require.NotNil(t, res.GetRefreshToken())
			},
		},
		{
			name: "Internal Error",
			req: &pb.LoginUserRequest{
				Username: user.Username,
				Password: password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(user, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, res *pb.LoginUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.Equal(t, codes.Internal, st.Code())
				require.True(t, ok)
			},
		},
		{
			name: "User Not Found",
			req: &pb.LoginUserRequest{
				Username: user.Username,
				Password: password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(user, db.ErrRecordNotFound)
			},
			checkResponse: func(t *testing.T, res *pb.LoginUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.Equal(t, codes.NotFound, st.Code())
				require.True(t, ok)
			},
		},
		{
			name: "Incorrect Password",
			req: &pb.LoginUserRequest{
				Username: user.Username,
				Password: "lmaoaturlife",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(t *testing.T, res *pb.LoginUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.Equal(t, codes.NotFound, st.Code())
				require.True(t, ok)
			},
		},
		{
			name: "Invalid Username",
			req: &pb.LoginUserRequest{
				Username: "user!name",
				Password: password,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0) // Should fail validation
			},
			checkResponse: func(t *testing.T, res *pb.LoginUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.Equal(t, codes.InvalidArgument, st.Code())
				require.True(t, ok)
			},
		},
		{
			name: "Not Email Validated",
			req: &pb.LoginUserRequest{
				Username: unverifiedUser.Username,
				Password: unverifedUserPW,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(unverifiedUser.Username)).
					Times(1).
					Return(unverifiedUser, nil)
			},
			checkResponse: func(t *testing.T, res *pb.LoginUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.Equal(t, codes.Unauthenticated, st.Code())
				require.True(t, ok)
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
			defer ctrl.Finish()
			taskDistributor := mockwk.NewMockTaskDistributor(tskCtrl)

			tc.buildStubs(mockStore)
			server := newTestServer(t, mockStore, taskDistributor)

			res, err := server.LoginUser(context.Background(), tc.req)
			tc.checkResponse(t, res, err)
		})
	}
}
