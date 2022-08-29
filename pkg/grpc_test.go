package pkg

import (
	"testing"

	grpcUtils "github.com/finiteloopme/goutils/pkg/grpc"
	grpcTest "github.com/finiteloopme/goutils/pkg/grpc/testing"
	"google.golang.org/grpc"
)

type TestHelloServer struct {
	// todo: fix hardcoding
	// userv1alpha1.UnimplementedHelloServiceServer
	grpcUtils.UnimplementedGRPCServer
}

func (quoteService TestHelloServer) Register(server *grpc.Server) {
	// userv1alpha1.RegisterHelloServiceServer(server, &TestHelloServer{})
}

func TestServer(t *testing.T) {
	var test TestHelloServer

	grpcTest.RegisterTestServer(test)
}
