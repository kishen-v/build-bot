package storage

import "io"

type ObjectInfo struct {
	ContentLength int64
	ContentType   string
}

type Storage interface {
	Download(bucket, object string) ([]byte, error)
	Upload(bucket string, object string, content []byte) error
	Get(bucket, object string) (io.ReadCloser, *ObjectInfo, error)
}