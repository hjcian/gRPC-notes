package serializer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hjcian/grpc-notes/pb"
	"gitlab.com/techschool/pcbook/serializer"
	"google.golang.org/protobuf/proto"

	"github.com/stretchr/testify/require"

	"github.com/hjcian/grpc-notes/sample"
)

func TestFileSerializer(t *testing.T) {
	t.Parallel()

	binFile := "../tmp/laptop.bin"
	jsonFile := "../tmp/laptop.json"

	// WRITE to file
	err := os.MkdirAll(filepath.Dir(binFile), 0755)
	// 0755 is rwx / r-x / r-x
	// drwxr-xr-x  3 maxcian  staff    96 12  6 15:50 tmp
	require.NoError(t, err)

	laptop1 := sample.NewLaptop()
	err = WriteProtobufToBinaryFile(laptop1, binFile)

	require.NoError(t, err)

	err = serializer.WriteProtobufToJSONFile(laptop1, jsonFile)
	require.NoError(t, err)

	// READ from file
	laptop2 := &pb.Laptop{}
	err = ReadProtobufFromBinaryFile(binFile, laptop2)
	require.NoError(t, err)

	require.True(t, proto.Equal(laptop1, laptop2))
}
