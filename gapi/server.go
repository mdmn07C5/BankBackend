package gapi

import (
	"fmt"

	db "github.com/mdmn07C5/bank/db/sqlc"
	"github.com/mdmn07C5/bank/pb"
	"github.com/mdmn07C5/bank/token"
	"github.com/mdmn07C5/bank/util"
)

// Server serves gRPC requests for our banking service
type Server struct {
	pb.UnimplementedBankRPCServiceServer
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
}

func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		store:      store,
		tokenMaker: tokenMaker,
		config:     config,
	}

	return server, nil
}
