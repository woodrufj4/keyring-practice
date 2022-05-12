package bbolt

import (
	"os"

	bolt "go.etcd.io/bbolt"
)

const (
	DefaultKeyringPath = "keyring.db"
)

// Config for using the bolt datastore
type Config struct {

	// Path is the file path to persist data
	Path string `json:"path"`

	// Filemode determines the access level to the database path
	Filemode os.FileMode `json:"filemode"`

	Options *bolt.Options `json:"options"`
}

// DefaultConfig provides sane defaults for the bolt datastore
func DefaultConfig() *Config {
	return &Config{
		Path:     DefaultKeyringPath,
		Filemode: 0600,
	}
}
