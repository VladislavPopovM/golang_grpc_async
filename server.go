package main

import (
	"context"
	"log"
	"net"
	"sync"
	"time"

	"micro/proto"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
)

type server struct {
	proto.UnimplementedMyServiceServer
}

func (s *server) MyMethod(ctx context.Context, req *proto.MyRequest) (*proto.MyResponse, error) {
	// Get a channel from the pool
	ch := channelPool.Get().(chan *proto.MyResponse)

	go func() {
		// Simulate some work by sleeping for 1 second
		time.Sleep(1 * time.Second)

		// Create a response message with a greeting
		resp := &proto.MyResponse{
			Message: "Hello, " + req.Name,
		}

		// Send the response message through the channel
		ch <- resp
	}()

	// Wait for the response message to be sent through the channel
	resp := <-ch

	// Put the channel back into the pool
	channelPool.Put(ch)

	return resp, nil
}

// Create a pool of channels
var channelPool = sync.Pool{
	New: func() interface{} {
		return make(chan *proto.MyResponse)
	},
}

func main() {
	// Create a new Fiber app
	app := fiber.New()

	// Create a listener on port 50051
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Create a new gRPC server
	grpcServer := grpc.NewServer()

	// Register the proto server with the gRPC server
	proto.RegisterMyServiceServer(grpcServer, &server{})

	// Start the gRPC server in a separate goroutine
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Start the Fiber app
	if err := app.Listen(":50052"); err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
}
