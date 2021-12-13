package minio

import (
	"bytes"
	"context"
	"sync"

	minioConn "github.com/fBloc/bloc-client-go/internal/conns/minio"
	"github.com/fBloc/bloc-client-go/internal/object_storage"

	"github.com/minio/minio-go/v7"
)

func init() {
	var _ object_storage.ObjectStorage = &ObjectStorageMinioRepository{}
}

type ObjectStorageMinioRepository struct {
	bucketName string
	client     *minio.Client
	sync.Mutex
}

func New(addresses []string, key, password string, bucketName string) *ObjectStorageMinioRepository {
	oSMR := &ObjectStorageMinioRepository{
		bucketName: bucketName,
		client:     minioConn.NewClient(addresses, key, password, bucketName),
	}
	return oSMR
}

func (oSMR *ObjectStorageMinioRepository) Set(key string, byteData []byte) error {
	objJsonIOReader := bytes.NewReader(byteData)
	_, err := oSMR.client.PutObject(
		context.Background(),
		oSMR.bucketName,
		key,
		objJsonIOReader,
		objJsonIOReader.Size(),
		minio.PutObjectOptions{ContentType: "application/octet-stream"})
	return err
}

func (oSMR *ObjectStorageMinioRepository) Get(key string) ([]byte, error) {
	reader, err := oSMR.client.GetObject(
		context.Background(), oSMR.bucketName, key, minio.GetObjectOptions{})
	if err != nil {
		return []byte{}, err
	}
	defer reader.Close()

	stat, err := reader.Stat()
	if err != nil {
		return []byte{}, err
	}
	data := make([]byte, stat.Size)
	reader.Read(data)
	return data, nil
}
