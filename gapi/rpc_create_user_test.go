package gapi

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	mockdb "github.com/mdmn07C5/bank/db/mock"
	db "github.com/mdmn07C5/bank/db/sqlc"
	"github.com/mdmn07C5/bank/pb"
	"github.com/mdmn07C5/bank/util"
	"github.com/mdmn07C5/bank/worker"
	mockwk "github.com/mdmn07C5/bank/worker/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type eqCreateUserTxParamsMatcher struct {
	arg           db.CreateUserTxParams
	nakedPassword string
	user          db.User
}

func (expected eqCreateUserTxParamsMatcher) Matches(x interface{}) bool {
	actualArg, ok := x.(db.CreateUserTxParams)
	if !ok {
		return false
	}
	err := util.CheckPassword(expected.nakedPassword, actualArg.HashedPassword)
	if err != nil {
		return false
	}
	expected.arg.HashedPassword = actualArg.HashedPassword
	if !reflect.DeepEqual(expected.arg.CreateUserParams, actualArg.CreateUserParams) {
		return false
	}

	err = actualArg.AfterCreation(expected.user)

	return err == nil
}

func (e eqCreateUserTxParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.nakedPassword)
}

func EqCreateUserTxParams(arg db.CreateUserTxParams, nakedPassword string, user db.User) gomock.Matcher {
	return eqCreateUserTxParamsMatcher{arg, nakedPassword, user}
}

func randomUser(t *testing.T) (user db.User, password string) {
	password = util.RandomString(8)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	user = db.User{
		Username:       util.RandomOwner(),
		Role:           util.DepositorRole,
		HashedPassword: hashedPassword,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}
	return

}

func randomVerifiedUser(t *testing.T) (user db.User, password string) {
	password = util.RandomString(8)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	user = db.User{
		Username:        util.RandomOwner(),
		Role:            util.DepositorRole,
		HashedPassword:  hashedPassword,
		FullName:        util.RandomOwner(),
		Email:           util.RandomEmail(),
		IsEmailVerified: true,
	}
	return

}

func requireBodyMatchUser(t *testing.T, body *bytes.Buffer, user db.User) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotUser db.User
	err = json.Unmarshal(data, &gotUser)

	require.NoError(t, err)
	require.Equal(t, user.Username, gotUser.Username)
	require.Equal(t, user.FullName, gotUser.FullName)
	require.Equal(t, user.Email, gotUser.Email)
	require.Empty(t, gotUser.HashedPassword)
}

func TestCreateUserAPI(t *testing.T) {
	user, password := randomUser(t)

	testCases := []struct {
		name          string
		req           *pb.CreateUserRequest
		buildStubs    func(mockStore *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor)
		checkResponse func(t *testing.T, res *pb.CreateUserResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.CreateUserRequest{
				Username: user.Username,
				Password: password,
				FullName: user.FullName,
				Email:    user.Email,
			},
			buildStubs: func(mockStore *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				userParams := db.CreateUserTxParams{
					CreateUserParams: db.CreateUserParams{
						Username: user.Username,
						FullName: user.FullName,
						Email:    user.Email,
					},
				}
				mockStore.EXPECT().
					CreateUserTx(gomock.Any(), EqCreateUserTxParams(userParams, password, user)).
					Times(1).
					Return(db.CreateUserTxResult{User: user}, nil)

				taskPayload := &worker.PayloadSendVerifyEmail{
					Username: user.Username,
				}
				taskDistributor.EXPECT().
					DistributeTaskSendVerifyEmail(gomock.Any(), taskPayload, gomock.Any()).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				createdUser := res.GetUser()
				require.Equal(t, user.Username, createdUser.Username)
				require.Equal(t, user.FullName, createdUser.FullName)
				require.Equal(t, user.Email, createdUser.Email)
			},
		},
		{
			name: "InternalError",
			req: &pb.CreateUserRequest{
				Username: user.Username,
				Password: password,
				FullName: user.FullName,
				Email:    user.Email,
			},
			buildStubs: func(mockStore *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				mockStore.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.CreateUserTxResult{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.Equal(t, codes.Internal, st.Code())
				require.True(t, ok)
			},
		},
		{
			name: "UserAlreadyExists",
			req: &pb.CreateUserRequest{
				Username: user.Username,
				Password: password,
				FullName: user.FullName,
				Email:    user.Email,
			},
			buildStubs: func(mockStore *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				mockStore.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.CreateUserTxResult{}, db.ErrUniqueViolation)
			},
			checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.Equal(t, codes.AlreadyExists, st.Code())
				require.True(t, ok)
			},
		},
		{
			name: "InvalidUsername",
			req: &pb.CreateUserRequest{
				Username: "invalid%username",
				Password: password,
				FullName: user.FullName,
				Email:    user.Email,
			},
			buildStubs: func(mockStore *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				mockStore.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.Equal(t, codes.InvalidArgument, st.Code())
				require.True(t, ok)
			},
		},
		{
			name: "InvalidPassword",
			req: &pb.CreateUserRequest{
				Username: user.Username,
				Password: "a",
				FullName: user.FullName,
				Email:    user.Email,
			},
			buildStubs: func(mockStore *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				mockStore.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.Equal(t, codes.InvalidArgument, st.Code())
				require.True(t, ok)
			},
		},
		{
			name: "InvalidFullname",
			req: &pb.CreateUserRequest{
				Username: user.Username,
				Password: password,
				FullName: "invalid^fullname",
				Email:    user.Email,
			},
			buildStubs: func(mockStore *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				mockStore.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.Equal(t, codes.InvalidArgument, st.Code())
				require.True(t, ok)
			},
		},
		{
			name: "InvalidEmail",
			req: &pb.CreateUserRequest{
				Username: user.Username,
				Password: password,
				FullName: user.FullName,
				Email:    "invalid email",
			},
			buildStubs: func(mockStore *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				mockStore.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.Equal(t, codes.InvalidArgument, st.Code())
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

			tc.buildStubs(mockStore, taskDistributor)

			server := newTestServer(t, mockStore, taskDistributor)
			res, err := server.CreateUser(context.Background(), tc.req)
			tc.checkResponse(t, res, err)
		})
	}
}
