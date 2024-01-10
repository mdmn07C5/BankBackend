package gapi

import (
	"context"

	db "github.com/mdmn07C5/bank/db/sqlc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/mdmn07C5/bank/pb"
)

func (server *Server) ListAccounts(ctx context.Context, req *pb.ListAccountsRequest) (*pb.ListAccountsResponse, error) {
	authPayload, err := server.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	arg := db.ListAccountsParams{
		Owner:  authPayload.Username,
		Limit:  req.GetPageSize(),
		Offset: (req.GetPageId() - 1) * req.GetPageSize(),
	}
	accounts, err := server.store.ListAccounts(ctx, arg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list accounts: %s", err)
	}

	accs := []*pb.Account{}
	for _, acc := range accounts {
		accs = append(accs, convertAccount(acc))
	}

	rsp := &pb.ListAccountsResponse{
		Accounts: accs,
	}
	return rsp, nil
}
