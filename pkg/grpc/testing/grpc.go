package testing

import (
	"context"
	"net"

	grpcUtils "github.com/finiteloopme/goutils/pkg/grpc"
	log "github.com/finiteloopme/goutils/pkg/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

type myInterface struct {
	grpcUtils.InterfaceGRPC
}

func (_myInterface *myInterface) dialer() func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	grpcUtils.InterfaceGRPC.Register(_myInterface, server)

	go func() {
		log.Info("Starting a gRPC server for testing")
		if err := server.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}

func RegisterTestServer(service grpcUtils.InterfaceGRPC) (conn *grpc.ClientConn, err error) {
	_myInterface := &myInterface{InterfaceGRPC: service}
	return grpc.DialContext(context.Background(), "", grpc.WithInsecure(), grpc.WithContextDialer(_myInterface.dialer()))
}
