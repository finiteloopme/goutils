package grpc

import (
	"context"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/kelseyhightower/envconfig"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type InterfaceGRPCWithHTTPHandler interface {
	RegisterHTTPHandler(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error)
	InterfaceGRPC
}

type UnimplementedGRPCWithHTTPHandler struct {
}

func (UnimplementedGRPCWithHTTPHandler) RegisterHTTPHandler(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error) {
	return fmt.Errorf("Placeholder.  Interface not implemented")
}

// Start HTTP Proxy to gRPC
func StartHTTPProxy(service InterfaceGRPCWithHTTPHandler, config GRPCConfig) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Register gRPC server endpoint
	// Note: Make sure the gRPC server is running properly and accessible
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	// err := gw.RegisterYourServiceHandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts)
	err := service.RegisterHTTPHandler(ctx, mux, config.GRPC_Host+":"+config.GRPC_Port, opts)
	if err != nil {
		return err
	}

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	return http.ListenAndServe(config.GRPC_Host+":"+config.HTTP_Port, mux)
}

// Start a gRPC Service with a REST endpoint too
func RunGRPCAndREST(service InterfaceGRPCWithHTTPHandler) error {
	// Start the gRPC server
	go RunGRPC(service)

	var config GRPCConfig
	envconfig.Process("gcp", &config)
	StartHTTPProxy(service, config)

	return nil
}
