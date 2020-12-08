package service_test

import (
	"context"
	"net"
	"testing"

	"github.com/hjcian/grpc-notes/sample"
	"github.com/hjcian/grpc-notes/serializer"

	"github.com/hjcian/grpc-notes/pb"
	"github.com/stretchr/testify/require"

	"github.com/hjcian/grpc-notes/service"
	"google.golang.org/grpc"
)

func startTestLaptopServer(t *testing.T, laptopstore service.LaptopStore) string {
	grpcServer := grpc.NewServer()

	laptopServer := service.NewLaptopServer(laptopstore)
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)

	listener, err := net.Listen("tcp", ":0") // random port number
	require.NoError(t, err)

	go grpcServer.Serve(listener)

	return listener.Addr().String() // only expose address string to client side
}

func newTestLaptopClient(t *testing.T, serverAddr string) pb.LaptopServiceClient {
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	require.NoError(t, err)

	return pb.NewLaptopServiceClient(conn)
}

func TestClientCreateLaptop(t *testing.T) {
	t.Parallel()

	// setup environment
	lpstore := service.NewInMemoryLaptopStore()
	addr := startTestLaptopServer(t, lpstore)
	lpClient := newTestLaptopClient(t, addr)

	// create a sample laptop
	laptop := sample.NewLaptop()
	expectedID := laptop.GetId()

	req := &pb.CreateLaptopRequest{
		Laptop: laptop,
	}

	// send one connection
	t.Run("send one connection", func(t *testing.T) {
		resp, err := lpClient.CreateLaptop(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, expectedID, resp.GetId())

	})
	t.Run("check that the laptop is saved to the store", func(t *testing.T) {
		// check that the laptop is saved to the store
		other, err := lpstore.Find(expectedID)
		require.NoError(t, err)
		require.NotNil(t, other)

		// check that the saved laptop is the same as the one we send
		requireSameLaptop(t, laptop, other)
	})
}

func requireSameLaptop(t *testing.T, laptop1 *pb.Laptop, laptop2 *pb.Laptop) {
	json1, err := serializer.ProtobufToJSON(laptop1)
	require.NoError(t, err)

	json2, err := serializer.ProtobufToJSON(laptop2)
	require.NoError(t, err)

	require.Equal(t, json1, json2)
}
