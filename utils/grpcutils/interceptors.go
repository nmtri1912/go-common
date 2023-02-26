package grpcutils

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func NewAuthenticatorUnaryInterceptor(clientId, clientKey string) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		callOpts ...grpc.CallOption,
	) error {
		newCtx := metadata.AppendToOutgoingContext(ctx, "client-id", clientId)
		newCtx = metadata.AppendToOutgoingContext(newCtx, "client-key", clientKey)
		return invoker(newCtx, method, req, reply, cc, callOpts...)
	}
}

func NewAuthenticatorStreamInterceptor(clientId, clientKey string) grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		callOpts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		newCtx := metadata.AppendToOutgoingContext(ctx, "client-id", clientId)
		newCtx = metadata.AppendToOutgoingContext(newCtx, "client-key", clientKey)
		return streamer(newCtx, desc, cc, method, callOpts...)
	}
}
