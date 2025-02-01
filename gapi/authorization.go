package gapi

import (
	"context"
	"fmt"
	"strings"

	"github.com/longtk26/simple_bank/token"
	"google.golang.org/grpc/metadata"
)

const (
	authorizationHeader = "authorization"
	authorizationType = "bearer"
)

func (server *Server) authorizeUser(ctx context.Context) (*token.Payload, error) {
	// Get metadata from context
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("metadata is not provided")
	}

	// Get authorization header from metadata
	values := md.Get(authorizationHeader)

	if len(values) == 0 {
		return nil, fmt.Errorf("authorization token is not provided")
	}

	authorizationHeader := strings.Fields(values[0])

	if len(authorizationHeader) < 2 {
		return nil, fmt.Errorf("authorization token is not provided")
	}

	authType := strings.ToLower(authorizationHeader[0])
	if authType != authorizationType {
		return nil, fmt.Errorf("authorization type is not provided")
	}

	accessToken := authorizationHeader[1]
	payload , err := server.tokenMaker.VerifyToken(accessToken)

	if err != nil {
		return nil, fmt.Errorf("token is invalid")
	}
	
	return payload, nil
}