package gapi

import (
	"context"

	db "github.com/mdmn07C5/bank/db/sqlc"
	"github.com/mdmn07C5/bank/util"

	"github.com/lib/pq"
	"github.com/mdmn07C5/bank/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) CreateAccount(ctx context.Context, req *pb.CreateAccountRequest) (*pb.CreateAccountResponse, error) {
	authPayload, err := server.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	currency := req.GetCurrency()
	if ok := util.IsSupportedCurrency(currency); !ok {
		return nil, status.Errorf(codes.InvalidArgument, "currency not supported: %s", currency)
	}

	arg := db.CreateAccountParams{
		Owner:    authPayload.Username,
		Currency: currency,
		Balance:  0,
	}

	account, err := server.store.CreateAccount(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				return nil, status.Errorf(codes.AlreadyExists, "account with that currency already exists: %s", err)
			}
		}
		return nil, status.Errorf(codes.Internal, "failed to create account: %s", err)
	}

	rsp := &pb.CreateAccountResponse{
		Account: convertAccount(account),
	}

	return rsp, nil
}
