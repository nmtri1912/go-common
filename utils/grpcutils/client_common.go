package grpcutils

import (
	"context"
	"crypto/tls"
	"log"
	"time"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	// grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
)

type cleanupFunc func()

type grpcConfig struct {
	sslEnable              bool
	target                 string
	keepAliveTimeInMinutes int
	clientId               string
	clientKey              string
}

type grpcCallConfig struct {
	deadlineSec int
}

func getGrpcConfig(service string) grpcConfig {
	return grpcConfig{
		sslEnable:              viper.GetBool(service + ".ssl-enabled"),
		target:                 viper.GetString(service + ".target"),
		keepAliveTimeInMinutes: viper.GetInt(service + ".keep-alive-time-in-minutes"),
		clientId:               viper.GetString(service + ".client-id"),
		clientKey:              viper.GetString(service + ".client-key"),
	}
}

func getGrpcCallConfig(service string) grpcCallConfig {
	return grpcCallConfig{
		deadlineSec: viper.GetInt(service + ".deadline-sec"),
	}
}

func CreateConnection(service string) (*grpc.ClientConn, cleanupFunc) {
	var credential grpc.DialOption
	grpcConfig := getGrpcConfig(service)

	if grpcConfig.sslEnable { // #nosec G402
		credential = grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true}))
	} else {
		credential = grpc.WithTransportCredentials(insecure.NewCredentials())
	}
	grpc.WithKeepaliveParams(keepalive.ClientParameters{})

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(viper.GetInt("grpc.connect-timeout-sec"))*time.Second)
	defer cancel()

	// grpc_prometheus.EnableClientHandlingTimeHistogram(grpc_prometheus.WithHistogramBuckets(prometheusutils.ReqDurBuckets))
	conn, err := grpc.DialContext(
		ctx,
		grpcConfig.target,
		credential,
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Timeout: time.Duration(grpcConfig.keepAliveTimeInMinutes) * time.Minute,
		}),
		grpc.WithChainUnaryInterceptor(
			otelgrpc.UnaryClientInterceptor(),
			NewAuthenticatorUnaryInterceptor(grpcConfig.clientId, grpcConfig.clientKey),
			// grpc_prometheus.UnaryClientInterceptor,
		),
		grpc.WithChainStreamInterceptor(
			otelgrpc.StreamClientInterceptor(),
			NewAuthenticatorStreamInterceptor(grpcConfig.clientId, grpcConfig.clientKey),
			// grpc_prometheus.StreamClientInterceptor,
		),
		grpc.WithBlock(),
	)

	if err != nil {
		log.Fatalf("Fail to dial %v: %v", service, err)
	}
	log.Println("Init grpc connection success", conn.Target())

	cleanup := func() {
		log.Print("Closing grpc connection")
		if err := conn.Close(); err != nil {
			log.Print("Close connection error", err)
		}
	}
	return conn, cleanup
}

func GetGrpcCallContext(ctx context.Context, service string) (context.Context, context.CancelFunc) {
	grpcCallConfig := getGrpcCallConfig(service)
	ctx, cancel := context.WithTimeout(ctx, time.Duration(grpcCallConfig.deadlineSec)*time.Second)
	return ctx, cancel
}
