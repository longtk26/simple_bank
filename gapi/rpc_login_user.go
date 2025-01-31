package gapi

import (
	"context"
	"database/sql"

	db "github.com/longtk26/simple_bank/db/sqlc"
	"github.com/longtk26/simple_bank/pb"
	"github.com/longtk26/simple_bank/util"
	"github.com/longtk26/simple_bank/val"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (server *Server) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {
	violations := validateLoginUserRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}
	
	user, err := server.store.GetUser(ctx, req.GetUsername())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
		}

		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	err = util.CheckPassword(req.GetPassword(), user.Password)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials: %v", err)
	}

	accessToken, payloadAccessToken, err := server.tokenMaker.CreateToken(
		user.Username,
		server.config.AccessTokenDuration,
	)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create access token: %v", err)
	}

	refreshToken, payloadRefreshToken, err := server.tokenMaker.CreateToken(
		user.Username,
		server.config.RefreshTokenDuration,
	)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create refresh token: %v", err)
	}

	mtdt := server.extractMetadata(ctx)

	session, err := server.store.CreateSession(
		ctx,
		db.CreateSessionParams {
			ID: payloadRefreshToken.ID,
			Username: user.Username,
			RefreshToken: refreshToken,
			ExpiresAt: payloadRefreshToken.ExpiredAt,
			UserAgent: mtdt.UserAgent,
			ClientIpAddress: mtdt.ClientIP,
			IsBlocked: false,
		},
	)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create session: %v", err)
	}

	rsp := &pb.LoginUserResponse{
		AccessToken: accessToken,
		AccessTokenExpires: timestamppb.New(payloadAccessToken.ExpiredAt),
		RefreshToken: refreshToken,
		SessionId: session.ID.String(),
		RefreshTokenExpires: timestamppb.New(payloadRefreshToken.ExpiredAt),
		User: convertUser(user),
	}

	return rsp, nil
}

func validateLoginUserRequest(req *pb.LoginUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolations("username", err))
	}

	if err := val.ValidatePassword(req.GetPassword()); err != nil {
		violations = append(violations, fieldViolations("password", err))
	}

	return violations
}