package grpc

import "google.golang.org/grpc/health/grpc_health_v1"

var HealthCheckPrefix = "/" + grpc_health_v1.Health_ServiceDesc.ServiceName
