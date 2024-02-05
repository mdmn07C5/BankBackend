package gapi

import (
	"context"
	"fmt"
	"time"

	db "github.com/mdmn07C5/bank/db/sqlc"
	"github.com/mdmn07C5/bank/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (server *Server) RenewAccessToken(ctx context.Context, req *pb.RenewAccessTokenRequest) (*pb.RenewAccessTokenResponse, error) {
	refreshPayload, err := server.tokenMaker.VerifyToken(req.GetRefreshToken())
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "refresh token is invalid: %s", err)
	}

	session, err := server.store.GetSession(ctx, refreshPayload.ID)
	if err != nil {
		if err == db.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "session not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to find session: %s", err)
	}

	if session.IsBlocked {
		err := fmt.Errorf("session blocked")
		return nil, unauthenticatedError(err)
	}

	if session.Username != refreshPayload.Username {
		err := fmt.Errorf("session user incorrect")
		return nil, unauthenticatedError(err)

	}

	if session.RefreshToken != req.GetRefreshToken() {
		err := fmt.Errorf("session token mismatched")
		return nil, unauthenticatedError(err)
	}

	if time.Now().After(session.ExpiresAt) {
		err := fmt.Errorf("session expired")
		return nil, unauthenticatedError(err)
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateToken(
		refreshPayload.Username,
		refreshPayload.Role,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create access token:%s", err)
	}

	rsp := &pb.RenewAccessTokenResponse{
		AccessToken: accessToken,
		ExpiresAt:   timestamppb.New(accessPayload.ExpiredAt),
	}

	return rsp, nil
}
