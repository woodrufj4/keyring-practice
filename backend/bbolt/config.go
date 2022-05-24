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

func (c *Config) Merge(b *Config) *Config {
	result := *c

	if b.Path != "" {
		result.Path = b.Path
	}

	if b.Filemode != 0 {
		result.Filemode = b.Filemode
	}

	if b.Options != nil {
		result.Options = b.Options
	}

	return &result
}

// DefaultConfig provides sane defaults for the bolt datastore
func DefaultConfig() *Config {
	return &Config{
		Path:     DefaultKeyringPath,
		Filemode: 0600,
	}
}
