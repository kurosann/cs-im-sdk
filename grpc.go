package csim

import (
	"context"

	gatewayv1 "github.com/kurosann/cs-im-sdk/gateway/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type ChannelGatewayClient = gatewayv1.ChannelGatewayServiceClient

func DialChannelGateway(ctx context.Context, endpoint, accountToken string) (ChannelGatewayClient, func(), error) {
	conn, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	closeFn := func() { _ = conn.Close() }
	return gatewayv1.NewChannelGatewayServiceClient(conn), closeFn, nil
}

func WithAccountToken(ctx context.Context, accountToken string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+accountToken)
}
