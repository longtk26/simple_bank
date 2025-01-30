package gapi

import (
	"context"

	"github.com/lib/pq"
	db "github.com/longtk26/simple_bank/db/sqlc"
	"github.com/longtk26/simple_bank/pb"
	"github.com/longtk26/simple_bank/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	hashedPassword, err := util.HashPassword(req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %v", err)
	}

	arg := db.CreateUserParams{
		Username: req.GetUsername(),
		Password: hashedPassword,
		FullName: req.GetFullName(),
		Email: req.GetEmail(),
	}

	user, err := server.store.CreateUser(ctx, arg)

	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok {
			switch pgErr.Code {
				case "unique_violation":
					return nil, status.Errorf(codes.AlreadyExists, "user already exists: %v", err)
			}
		}
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return &pb.CreateUserResponse{
		User: convertUser(user),
	}, nil
}

