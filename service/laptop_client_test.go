package service_test

import (
	"context"
	"io"
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

func TestClientSearchLaptop(t *testing.T) {
	t.Parallel()

	store := service.NewInMemoryLaptopStore()

	expectedIDs := make(map[string]bool)

	for i := 0; i < 6; i++ {
		laptop := sample.NewLaptop()

		switch i {
		case 0:
			// Case 0: unmatched laptop with a too high price.
			laptop.PriceUsd = 2500
		case 1:
			// Case 1: unmatched because it has only 2 cores.
			laptop.Cpu.NumberCores = 2
		case 2:
			// Case 2: doesn’t match because the min frequency is too low.
			laptop.Cpu.MinGhz = 2.0
		case 3:
			// Case 3: doesn’t match since it has only 4 GB of RAM.
			laptop.Ram = &pb.Memory{Value: 4096, Unit: pb.Memory_MEGABYTE}
		case 4:
			// Case 4 + 5: matched.
			laptop.PriceUsd = 1999
			laptop.Cpu.NumberCores = 4
			laptop.Cpu.MinGhz = 2.5
			laptop.Cpu.MaxGhz = laptop.Cpu.MinGhz + 2.0
			laptop.Ram = &pb.Memory{Value: 16, Unit: pb.Memory_GIGABYTE}
			expectedIDs[laptop.Id] = true
		case 5:
			// Case 4 + 5: matched.
			laptop.PriceUsd = 2000
			laptop.Cpu.NumberCores = 6
			laptop.Cpu.MinGhz = 2.8
			laptop.Cpu.MaxGhz = laptop.Cpu.MinGhz + 2.0
			laptop.Ram = &pb.Memory{Value: 64, Unit: pb.Memory_GIGABYTE}
			expectedIDs[laptop.Id] = true
		}

		err := store.Save(laptop)
		require.NoError(t, err)
	}

	filter := &pb.Filter{
		MaxPriceUsd: 2000,
		MinCpuCores: 4,
		MinCpuGhz:   2.2,
		MinRam:      &pb.Memory{Value: 8, Unit: pb.Memory_GIGABYTE},
	}
	serverAddress := startTestLaptopServer(t, store)
	client := newTestLaptopClient(t, serverAddress)

	req := &pb.SearchLaptopRequest{Filter: filter}

	stream, err := client.SearchLaptop(context.Background(), req)

	require.NoError(t, err)
	found := 0

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}

		require.NoError(t, err)
		require.Contains(t, expectedIDs, res.GetLaptop().GetId())
		found += 1
	}

	require.Equal(t, len(expectedIDs), found)
}
