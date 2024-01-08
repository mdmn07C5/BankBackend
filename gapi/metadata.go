package gapi

import (
	"context"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const (
	grpcGatewayUserAgentHeader = "grpcgateway-user-agent"
	userAgentHeader            = "user-agent"
	xForwardedForHeader        = "x-forwarded-host"
)

type Metadata struct {
	UserAgent string
	ClientIP  string
}

func (server *Server) extractMetadata(ctx context.Context) *Metadata {
	meta := &Metadata{}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if userAgents := md.Get(grpcGatewayUserAgentHeader); len(userAgents) > 0 {
			meta.UserAgent = userAgents[0]
		}

		if userAgents := md.Get(userAgentHeader); len(userAgents) > 0 {
			meta.UserAgent = userAgents[0]
		}

		if clientIP := md.Get(xForwardedForHeader); len(clientIP) > 0 {
			meta.ClientIP = clientIP[0]
		}

	}

	if p, ok := peer.FromContext(ctx); ok {
		meta.ClientIP = p.Addr.String()
	}

	return meta
}
