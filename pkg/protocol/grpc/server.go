package grpc

import (
	"context"
	"log"
	"net"
	"os"

	v1 "github.com/devararishivian/go-grpc/pkg/api/v1"
	"google.golang.org/grpc"
)

// RunServer runs gRPC service to publish Todo service
func RunServer(ctx context.Context, v1API v1.TodoServiceServer, port string) error {
	listen, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	// Register service
	server := grpc.NewServer()
	v1.RegisterTodoServiceServer(server, v1API)

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	go func() {
		for range c {
			log.Println("shutting down gRPC server...")
			server.GracefulStop()

			<-ctx.Done()
		}
	}()

	// Start gRPC server
	log.Println("starting gRPC server...")
	return server.Serve(listen)
}
