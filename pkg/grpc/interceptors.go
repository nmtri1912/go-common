package grpc

import (
	"context"
	"strings"

	"github.com/nmtri1912/go-common/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const ClientIdMetadataKey = "client-id"
const ClientKeyMetadataKey = "client-key"

func NewAuthenUnaryServerInterceptor(clients map[string]string, methodClients map[string][]string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if strings.HasPrefix(info.FullMethod, HealthCheckPrefix) {
			//skip for health check
			return handler(ctx, req)
		}
		requestMetadata, _ := metadata.FromIncomingContext(ctx)
		clientId, clientKey := requestMetadata.Get(ClientIdMetadataKey), requestMetadata.Get(ClientKeyMetadataKey)
		if len(clientId) <= 0 || len(clientKey) <= 0 {
			return nil, status.Error(codes.Unauthenticated, "client-id or client-key not present in metadata")
		}
		serverClientKey, exist := clients[clientId[0]]
		if !exist {
			return nil, status.Error(codes.Unauthenticated, "client-id not found")
		}
		if serverClientKey != clientKey[0] {
			return nil, status.Error(codes.Unauthenticated, "client-key mismatch")
		}
		method := strings.ToLower(info.FullMethod)
		allowedClients, exist := methodClients[method]
		if exist && !contains(allowedClients, clientId[0]) {
			return nil, status.Error(codes.Unauthenticated, "client-id not allowed")
		}
		resp, err := handler(ctx, req)
		return resp, err
	}
}

func NewLoggingUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if strings.HasPrefix(info.FullMethod, HealthCheckPrefix) {
			//skip for health check
			return handler(ctx, req)
		}
		requestMetadata, _ := metadata.FromIncomingContext(ctx)
		logger.Ctx(ctx).Info("Request: ",
			zap.String("method", info.FullMethod),
			zap.Reflect("metadata", requestMetadata),
			zap.Reflect("request", req),
		)
		resp, err := handler(ctx, req)
		if err != nil {
			code, reason := ExtractCodeAndReasonFromError(err)
			if code == codes.Internal {
				logger.Ctx(ctx).Error("Response: ", zap.Error(err))
			} else {
				logger.Ctx(ctx).Warn("Response: ", zap.String("reason", reason))
			}
		} else {
			logger.Ctx(ctx).Info("Response: ", zap.Reflect("response", resp))
		}
		return resp, err
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func NewRecoverUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
		panicked := true
		defer func() {
			if r := recover(); r != nil || panicked {
				err = status.Errorf(codes.Internal, "%v", r)
				logger.Ctx(ctx).Error("Response: ", zap.Error(err))
			}
		}()
		resp, err := handler(ctx, req)
		panicked = false
		return resp, err
	}

}

func ExtractCodeAndReasonFromError(err error) (codes.Code, string) {
	status, ok := status.FromError(err)
	if !ok {
		return codes.Internal, err.Error()
	}
	if len(status.Details()) <= 0 {
		return status.Code(), err.Error()
	}
	errInfo, ok := status.Details()[0].(*errdetails.ErrorInfo)
	if !ok {
		return status.Code(), err.Error()
	}
	return status.Code(), errInfo.GetReason()
}
