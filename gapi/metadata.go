package gapi

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const (
	grpcGatewayUserAgentHeader = "grpcgateway-user-agent"
	xForwardedForHeader = "x-forwarded-for"
	userAgentHeader = "user-agent"
)

type Metadata struct {
	UserAgent string
	ClientIP string
}

func (server *Server) extractMetadata(ctx context.Context) *Metadata {
	mtdt := &Metadata{}

	if md, ok := metadata.FromIncomingContext(ctx); ok {

		if userAgents := md.Get(grpcGatewayUserAgentHeader); len(userAgents) > 0 {
			mtdt.UserAgent = userAgents[0]
		}

		if userAgents := md.Get(userAgentHeader); len(userAgents) > 0 {
			mtdt.UserAgent = userAgents[0]
		}

		if clientIPs := md.Get(xForwardedForHeader); len(clientIPs) > 0 {
			mtdt.ClientIP = clientIPs[0]
		}

	}

	peerInfo, ok := peer.FromContext(ctx)
	if ok {
		log.Printf("peer info: %v", peerInfo)

		if addr, ok := peerInfo.Addr.(*net.TCPAddr); ok {
			mtdt.ClientIP = addr.IP.String()
		}
	}

	return mtdt
}