package storage

//go:generate mockgen -source=server.go -destination=server_mock.go -package=storage

import (
	"context"
	"io"
)

// Server - interface of an abstract storage server
type Server interface {
	ID() int // Server node ID = array index for simplicity
	Store(ctx context.Context, name string, size uint64, obj io.Reader) error
	Get(ctx context.Context, name string) (io.ReadCloser, uint64, error)
	Remove(ctx context.Context, name string) error
	FreeSpace() float64 // free space percentage
}
