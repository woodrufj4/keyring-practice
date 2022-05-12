package backend

import (
	"context"
)

type BackendType string

const (

	// FileBackend persists data directly to disk.
	//
	// This uses bbolt util to persis to disk.
	// https://github.com/etcd-io/bbolt
	FileBackend BackendType = "file"
)

type Backend interface {

	// Type reports the type of backend
	Type() BackendType

	// Setup is used to setup the backend based on the provided
	// configuration
	Setup(context.Context) error

	// Cleanup is invoked during an unmounting or shutdown to allow
	// it to hanlde cleanup like connection closing or releasing
	// resource handles.
	Cleanup(context.Context) error

	// Persists data to the backend
	Put(ctx context.Context, path string, entries []*BackendEntry) error

	// Get retrieves the backend entry at provided path.
	Get(ctx context.Context, path string) ([]*BackendEntry, error)

	// List reports the existing paths.
	List(ctx context.Context) ([]string, error)

	// Delete removes all secrets at the given path
	Delete(ctx context.Context, path string) error
}

type BackendEntry struct {
	Key   string
	Value []byte
}
