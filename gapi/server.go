package gapi

import (
	"fmt"

	db "github.com/longtk26/simple_bank/db/sqlc"
	"github.com/longtk26/simple_bank/pb"
	"github.com/longtk26/simple_bank/token"
	"github.com/longtk26/simple_bank/util"
)

type Server struct {
	pb.UnimplementedSimpleBankServer
	config util.Config
	store db.Store
	tokenMaker token.IMaker
}

// New server creates a new gRPC server 
func NewServer (config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)

	if err != nil {
		return nil, fmt.Errorf("Cannot create token maker %w", err)
	}

	server := &Server{
		config: config,
		store: store,
		tokenMaker: tokenMaker,
	}

	return server, nil
}