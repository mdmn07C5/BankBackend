package gapi

import (
	db "github.com/mdmn07C5/bank/db/sqlc"
	"github.com/mdmn07C5/bank/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func convertUser(user db.User) *pb.User {
	return &pb.User{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: timestamppb.New(user.PasswordChangedAt),
		CreatedAt:         timestamppb.New(user.CreatedAt),
	}
}

func convertAccount(account db.Account) *pb.Account {
	return &pb.Account{
		Id:        account.ID,
		Owner:     account.Owner,
		Balance:   account.Balance,
		Currency:  account.Currency,
		CreatedAt: timestamppb.New(account.CreatedAt),
	}
}

func convertEntry(entry db.Entry) *pb.Entry {
	return &pb.Entry{
		Id:        entry.ID,
		AccountId: entry.AccountID,
		Amount:    entry.Amount,
		CreatedAt: timestamppb.New(entry.CreatedAt),
	}
}

func convertTransferResult(transferResult db.TransferTxResult) *pb.TransferResponse {
	transfer := &pb.Transfer{
		Id:            transferResult.Transfer.ID,
		FromAccountId: transferResult.Transfer.FromAccountID,
		ToAccountId:   transferResult.Transfer.ToAccountID,
		Amount:        transferResult.Transfer.Amount,
		CreatedAt:     timestamppb.New(transferResult.Transfer.CreatedAt),
	}

	return &pb.TransferResponse{
		Transfer:    transfer,
		FromAccount: convertAccount(transferResult.FromAccount),
		ToAccount:   convertAccount(transferResult.ToAccount),
		FromEntry:   convertEntry(transferResult.FromEntry),
		ToEntry:     convertEntry(transferResult.ToEntry),
	}
}
