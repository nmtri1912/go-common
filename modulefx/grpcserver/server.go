package grpcserver

import (
	"context"
	"fmt"
	"log"
	"net"

	grpc_util "github.com/nmtri1912/go-common/pkg/grpc"

	"github.com/spf13/viper"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type GrpcService struct {
	ServiceDesc          *grpc.ServiceDesc
	ServiceImpl          interface{}
	Clients              map[string]string
	AllowedMethodClients map[string][]string
}

func StartGrpcServer(lifecycle fx.Lifecycle, service *GrpcService) error {
	port := viper.GetInt("grpc.port")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpc_util.NewRecoverUnaryServerInterceptor(),
			grpc_util.NewTracingUnaryServerInterceptor(),
			grpc_util.NewLoggingUnaryServerInterceptor(),
			grpc_util.NewAuthenUnaryServerInterceptor(service.Clients, service.AllowedMethodClients),
		),
	)
	//health check
	grpc_health_v1.RegisterHealthServer(grpcServer, health.NewServer())
	//actual service
	grpcServer.RegisterService(service.ServiceDesc, service.ServiceImpl)

	lifecycle.Append(fx.Hook{OnStart: func(ctx context.Context) error {
		go func() {
			log.Println("gRPC server starting on port: ", port)
			err := grpcServer.Serve(lis)
			if err != nil {
				log.Println("grpcServer.Serve has error: ", err.Error())
			}
		}()
		return nil
	}, OnStop: func(c context.Context) error {
		log.Println("gRPC server Shutting down...")
		grpcServer.GracefulStop()
		log.Println("gRPC server Shutted down")
		return nil
	}})
	return nil
}
