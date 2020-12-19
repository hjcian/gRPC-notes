package service_test

import (
	"context"
	"testing"

	"github.com/hjcian/grpc-notes/service"

	"github.com/hjcian/grpc-notes/pb"
	"github.com/hjcian/grpc-notes/sample"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServerCreateLaptop(t *testing.T) {
	t.Parallel()

	laptopNoID := sample.NewLaptop()
	laptopNoID.Id = ""

	laptopInvalidID := sample.NewLaptop()
	laptopInvalidID.Id = "invalid-uuid"

	laptopDuplicateID := sample.NewLaptop()
	storeDuplicateID := service.NewInMemoryLaptopStore()
	err := storeDuplicateID.Save(laptopDuplicateID)
	require.NoError(t, err)

	testImageFolder := "../tmp"
	imageStore := service.NewDiskImageStore(testImageFolder)

	testCases := []struct {
		name       string
		laptop     *pb.Laptop
		store      service.LaptopStore
		imageStore service.ImageStore
		code       codes.Code
	}{
		{
			name:       "success_with_id",
			laptop:     sample.NewLaptop(),
			store:      service.NewInMemoryLaptopStore(),
			imageStore: imageStore,
			code:       codes.OK,
		},
		{
			name:       "success_no_id",
			laptop:     laptopNoID,
			store:      service.NewInMemoryLaptopStore(),
			imageStore: imageStore,
			code:       codes.OK,
		},
		{
			name:       "failure_invalid_id",
			laptop:     laptopInvalidID,
			store:      service.NewInMemoryLaptopStore(),
			imageStore: imageStore,
			code:       codes.InvalidArgument,
		},
		{
			name:       "failure_duplicate_id",
			laptop:     laptopDuplicateID,
			store:      storeDuplicateID,
			imageStore: imageStore,
			code:       codes.AlreadyExists,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := &pb.CreateLaptopRequest{
				Laptop: tc.laptop,
			}

			server := service.NewLaptopServer(tc.store, tc.imageStore)

			res, err := server.CreateLaptop(context.Background(), req)

			if tc.code == codes.OK {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.NotEmpty(t, res.GetId())
				if len(tc.laptop.GetId()) > 0 {
					require.Equal(t, tc.laptop.GetId(), res.GetId())
				}
			} else {
				require.Error(t, err)
				require.Nil(t, res)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, tc.code, st.Code())
			}
		})
	}
}
