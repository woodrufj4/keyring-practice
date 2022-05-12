package bbolt

import (
	"context"
	"fmt"

	"github.com/woodrufj4/keyring-practice/backend"
	bolt "go.etcd.io/bbolt"
)

const (
	ErrMissingConfig = "missing configuration"
	ErrMissingPath   = "missing file path"
)

type BoltBackend struct {
	config *Config
	db     *bolt.DB
}

func NewBoltBackend(config *Config) (backend.Backend, error) {

	if config == nil {
		return nil, fmt.Errorf(ErrMissingConfig)
	}

	if config.Path == "" {
		return nil, fmt.Errorf(ErrMissingPath)
	}

	// @TODO: validate filemode

	return &BoltBackend{
		config: config,
	}, nil

}

// Type reports the type of backend
func (b BoltBackend) Type() backend.BackendType {
	return backend.FileBackend
}

// Setup initializes the file datastore based on the provided
// configuration
//
// The caller is responsible for calling Cleanup() to
func (b *BoltBackend) Setup(context.Context) error {

	db, err := bolt.Open(b.config.Path, b.config.Filemode, b.config.Options)

	if err != nil {
		return err
	}

	b.db = db

	return nil
}

// Cleanup is invoked during an unmounting or shutdown to allow
// it to hanlde cleanup like connection closing or releasing
// resource handles.
func (b *BoltBackend) Cleanup(context.Context) error {

	if b.db != nil {
		return b.db.Close()
	}

	return nil
}

// Put persists data to the backend at the provided path.
func (b *BoltBackend) Put(ctx context.Context, path string, entries []*backend.BackendEntry) error {

	tx, err := b.db.Begin(true)

	if err != nil {
		return err
	}

	bucket, err := tx.CreateBucketIfNotExists([]byte(path))

	if err != nil {
		rbErr := tx.Rollback()

		if rbErr != nil {
			return fmt.Errorf("failed to create bucket: %s. rollback failed: %s", err.Error(), rbErr.Error())
		}

		return fmt.Errorf("failed to created bucket: %s", err.Error())
	}

	for _, entry := range entries {

		err = bucket.Put([]byte(entry.Key), entry.Value)

		if err != nil {

			rbErr := tx.Rollback()

			if rbErr != nil {
				return fmt.Errorf("failed to put entry at path '%s': %s. rollback failed: %s", path, err.Error(), rbErr.Error())
			}

			return fmt.Errorf("failed to put entry at path '%s': %s", path, err.Error())
		}

	}

	return tx.Commit()

}

func (b *BoltBackend) Get(ctx context.Context, path string) ([]*backend.BackendEntry, error) {

	tx, err := b.db.Begin(false)

	if err != nil {
		return nil, err
	}

	bucket := tx.Bucket([]byte(path))

	if bucket == nil {
		return nil, tx.Rollback()
	}

	entries := make([]*backend.BackendEntry, 0)

	err = bucket.ForEach(func(k, v []byte) error {
		entries = append(entries, &backend.BackendEntry{
			Key:   string(k),
			Value: v,
		})
		return nil
	})

	if err != nil {
		rbErr := tx.Rollback()

		if rbErr != nil {
			return nil, fmt.Errorf("failed to get entry at path '%s': %s. rollback failed: %s", path, err.Error(), rbErr.Error())
		}

		return nil, fmt.Errorf("failed to get entry at path '%s': %s", path, err.Error())
	}

	// Read-only transactions must be rolled back and not commited
	err = tx.Rollback()

	if err != nil {
		return nil, fmt.Errorf("failed to rollback transaction: %s", err.Error())
	}

	return entries, nil
}

func (b *BoltBackend) List(ctx context.Context) ([]string, error) {
	tx, err := b.db.Begin(false)

	if err != nil {
		return nil, err
	}

	paths := []string{}

	err = tx.ForEach(func(name []byte, b *bolt.Bucket) error {
		paths = append(paths, string(name))
		return nil
	})

	if err != nil {
		rbErr := tx.Rollback()

		if rbErr != nil {
			return nil, fmt.Errorf("failed to list root paths: %s. rollback failed: %s", err.Error(), rbErr.Error())
		}

		return nil, fmt.Errorf("failed to list root paths: %s", err.Error())
	}

	err = tx.Rollback()

	if err != nil {
		return nil, fmt.Errorf("failed to rollback transaction: %s", err.Error())
	}

	return paths, nil
}

func (b *BoltBackend) Delete(ctx context.Context, path string) error {
	tx, err := b.db.Begin(true)

	if err != nil {
		return err
	}

	err = tx.DeleteBucket([]byte(path))

	if err != nil {
		rbErr := tx.Rollback()

		if rbErr != nil {
			return fmt.Errorf("failed to delete secrets at path '%s': %s. rollback failed: %s", path, err.Error(), rbErr.Error())
		}

		return fmt.Errorf("failed to delete secrets at path '%s': %s", path, err.Error())
	}

	return tx.Commit()

}
