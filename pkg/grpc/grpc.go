// Simple utility to start a gRPC Service
// Consumer should:
// 1. Implement `InterfaceGRPC`
// 2. Call: RunGRPC()
package grpc

import (
	"fmt"
	"net"

	log "github.com/finiteloopme/goutils/pkg/log"
	"google.golang.org/grpc"
)

type InterfaceGRPC interface {
	Register(*grpc.Server)
}

type UnimplementedGRPCServer struct {
}

func (UnimplementedGRPCServer) Register(server *grpc.Server) {
	return
}

// Config for gRPC Server
type GRPCConfig struct {
	// Set env variable GCP_GRPC_HOST. Default value is 0.0.0.0
	GRPC_Host string `default:"0.0.0.0"`
	// Set env variable GCP_GRPC_PORT. Default value is 8080
	GRPC_Port string `default:"8081"`
}

// Start the gRPC server
func RunGRPC(service InterfaceGRPC) error {
	var config GRPCConfig
	listenOn := config.GRPC_Host + ":" + config.GRPC_Port
	listener, err := net.Listen("tcp", listenOn)
	if err != nil {
		return fmt.Errorf("Failed to listen on: %w", err)
	}
	server := grpc.NewServer()
	service.Register(server)
	log.Info("Listening on: " + listenOn)
	if err := server.Serve(listener); err != nil {
		return fmt.Errorf("Failed to start gRPC server: %w", err)
	}
	return nil
}
