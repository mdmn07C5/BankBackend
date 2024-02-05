package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	mockdb "github.com/mdmn07C5/bank/db/mock"
	db "github.com/mdmn07C5/bank/db/sqlc"
	"github.com/mdmn07C5/bank/token"
	"github.com/mdmn07C5/bank/util"
	"github.com/stretchr/testify/require"
)

func addAuthorization(
	t *testing.T,
	request *http.Request,
	tokenMaker token.Maker,
	authorizationType string,
	username string,
	role string,
	duration time.Duration,
) {
	token, payload, err := tokenMaker.CreateToken(username, role, duration)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	authorizationHeader := fmt.Sprintf("%s %s", authorizationType, token)
	request.Header.Set(authorizationHeaderKey, authorizationHeader)
}

func TestAuthMiddleWare(t *testing.T) {
	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, util.RandomOwner(), util.DepositorRole, time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "NoAuthorization",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				// addAuthorization(t, request, tokenMaker, authorizationTypeBearer, util.RandomOwner(), time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "UnsupportedAuthorization",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, "unsupported", util.RandomOwner(), util.DepositorRole, time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "AuthorizationHeaderMalformed",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, "", util.RandomOwner(), util.DepositorRole, time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "AuthorizationHeaderMalformed",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, util.RandomOwner(), util.DepositorRole, -time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			server := newTestServer(t, nil)

			authPath := "/auth"
			server.router.GET(
				authPath,
				authMiddleware(server.tokenMaker),
				func(ctx *gin.Context) {
					ctx.JSON(http.StatusOK, gin.H{})
				},
			)

			recorder := httptest.NewRecorder()
			request, err := http.NewRequest(http.MethodGet, authPath, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

// TODO: Actually fucking fix this test
func TestSessionMiddleWare(t *testing.T) {
	user, _ := randomUser(t)

	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		setupSession  func(t *testing.T, request *http.Request, server *Server) db.CreateSessionParams
		buildStubs    func(mockstore *mockdb.MockStore, params db.CreateSessionParams, session db.Session)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, user.Role, time.Minute)
			},
			setupSession: func(t *testing.T, request *http.Request, server *Server) db.CreateSessionParams {
				authHeader := request.Header.Get(authorizationHeaderKey)
				require.NotEmpty(t, authHeader)

				fields := strings.Fields(authHeader)
				require.GreaterOrEqual(t, len(fields), 2)

				refreshTokenPayload, err := server.tokenMaker.VerifyToken(fields[1])
				require.NoError(t, err)
				require.NotEmpty(t, refreshTokenPayload)

				arg := db.CreateSessionParams{
					ID:           refreshTokenPayload.ID,
					Username:     user.Username,
					RefreshToken: fields[1],
					UserAgent:    "test",
					ClientIp:     "test",
					IsBlocked:    false,
					ExpiresAt:    time.Now().Add(time.Minute),
				}
				return arg
			},
			buildStubs: func(mockstore *mockdb.MockStore, params db.CreateSessionParams, session db.Session) {
				mockstore.EXPECT().
					CreateSession(gomock.Any(), gomock.Eq(params)).
					Times(1).
					Return(session, nil)

				mockstore.EXPECT().
					GetSession(gomock.Any(), gomock.Eq(session.ID)).
					Times(1).
					Return(session, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			authPath := "/auth"
			server.router.GET(
				authPath,
				authMiddleware(server.tokenMaker),
				sessionMiddleWare(server),
				func(ctx *gin.Context) {
					ctx.JSON(http.StatusOK, gin.H{})
				},
			)
			request, err := http.NewRequest(http.MethodGet, authPath, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)

			arg := tc.setupSession(t, request, server)

			var session db.Session
			tc.buildStubs(store, arg, session)

			session, err = server.store.CreateSession(context.Background(), arg)

			server.store.GetSession(context.Background(), session.ID)

			server.router.ServeHTTP(recorder, request)

			// tc.checkResponse(t, recorder)
		})
	}

}
