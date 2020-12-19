package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/hjcian/grpc-notes/pb"
	"google.golang.org/grpc"

	"github.com/hjcian/grpc-notes/service"
)

func main() {
	port := flag.Int("port", 0, "ther server port")
	flag.Parse()
	log.Printf("start server on port %d", *port)

	grpcServer := grpc.NewServer()
	lpServer := service.NewLaptopServer(
		service.NewInMemoryLaptopStore(),
		service.NewDiskImageStore("img"),
	)
	pb.RegisterLaptopServiceServer(grpcServer, lpServer)

	addr := fmt.Sprintf("0.0.0.0:%d", *port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("cannot listen on addr %s: %s", addr, err)
	}

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatalf("cannot start gRPC server: %s", err)
	}

}
