// Simple utility to start a gRPC Service
// Consumer should:
// 1. Implement `InterfaceGRPC`
// 2. Call: RunGRPC()
package grpc

import (
	"fmt"
	"net"

	log "github.com/finiteloopme/goutils/pkg/log"
	"github.com/kelseyhightower/envconfig"
	"google.golang.org/grpc"
)

type InterfaceGRPC interface {
	Register(*grpc.Server)
}

type UnimplementedGRPCServer struct {
}

func (UnimplementedGRPCWithHTTPHandler) Register(server *grpc.Server) error {
	return fmt.Errorf("Placeholder.  Interface not implemented")
}

// Config for gRPC Server
type GRPCConfig struct {
	// Set env variable GCP_GRPC_HOST. Default value is 0.0.0.0
	GRPC_Host string `default:"0.0.0.0"`
	// Set env variable GCP_GRPC_PORT. Default value is 8080
	GRPC_Port string `default:"8080"`
	// Set env variable GCP_HTTP_PORT. Default value is 8090
	HTTP_Port string `default:"8090"`
}

// Start the gRPC server
func RunGRPC(service InterfaceGRPC) error {
	var config GRPCConfig
	envconfig.Process("gcp", &config)
	listenOn := config.GRPC_Host + ":" + config.GRPC_Port
	listener, err := net.Listen("tcp", listenOn)
	if err != nil {
		return fmt.Errorf("Failed to listen on: %w", err)
	}
	server := grpc.NewServer()
	service.Register(server)
	log.Info("Starting gRPC service on: " + listenOn)
	if err := server.Serve(listener); err != nil {
		return fmt.Errorf("Failed to start gRPC server: %w", err)
	}
	return nil
}
