package gapi

import (
	"context"
	"database/sql"
	"errors"

	db "github.com/mdmn07C5/bank/db/sqlc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/mdmn07C5/bank/pb"
)

func (server *Server) TransferFunds(ctx context.Context, req *pb.TransferRequest) (*pb.TransferResponse, error) {
	fromAccount, err := server.validAccount(ctx, req.GetFromAccountId(), req.GetCurrency())
	if err != nil {
		return nil, err
	}

	authPayload, err := server.authorizeUser(ctx)
	if fromAccount.Owner != authPayload.Username {
		err := errors.New("from account does not belong to the authenticated user")
		return nil, unauthenticatedError(err)
	}

	// validate toAccount
	if _, err := server.validAccount(ctx, req.GetToAccountId(), req.Currency); err != nil {
		return nil, err
	}

	arg := db.TransferTxParams{
		FromAccountID: req.GetFromAccountId(),
		ToAccountID:   req.GetToAccountId(),
		Amount:        req.GetAmount(),
	}

	result, err := server.store.TransferTx(ctx, arg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "transfer failed: %s", err)
	}

	return convertTransferResult(result), nil
}

func (server *Server) validAccount(ctx context.Context, accountID int64, currency string) (db.Account, error) {
	account, err := server.store.GetAccount(ctx, accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			return account, status.Errorf(codes.NotFound, "account not found:%s", err)
		}
		return account, status.Errorf(codes.Internal, "failed to validate account:%s", err)
	}

	if account.Currency != currency {
		return account, status.Errorf(codes.InvalidArgument, "account [%d] currency mismatch: %s vs %s", accountID, account.Currency, currency)
	}

	return account, nil
}
