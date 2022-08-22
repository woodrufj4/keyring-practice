package command

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mitchellh/cli"
	"github.com/woodrufj4/keyring-practice/backend"
	"github.com/woodrufj4/keyring-practice/backend/bbolt"
	"github.com/woodrufj4/keyring-practice/internal"
)

type KVPutCommand struct {
	ui cli.Ui
}

func (kv KVPutCommand) Synopsis() string {
	return "Stores and encrypts key-value pairs at the provided path"
}

func (kv KVPutCommand) Help() string {
	helpText := `
Usage: keying put [options] <path> [data]

  This stores and encrypts a key-value pairs at the given path.
  The path is used to lookup the key-value pairs during decryption.

  There can be any number of key-value pairs. For example:

    $ keyring put secret/foo bar=baz key=secret

  The data can also be consumed from a file on disk by prefixing with the "@"
  symbol. For example:

    $ keyring put secret/foo @data.json

  Options:

    -root-token=<string>
      The root token to use for encryption.
      If not provided here, the '%s' environment
      variable will be used.

  Backend Options:

    -backend-type=<string>
      The type of backend to use.
      Currently, only the 'file' type backend is supported,
      and is also the default. 


    File Backend Options:

      -filepath=<string>
        The file path where your secrets will be persisted to disc.
`
	return fmt.Sprintf(helpText, internal.DefaultEnvRootToken)
}

func (kv *KVPutCommand) Run(args []string) int {

	defaultCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	config, fs, err := kv.readConfig(args)

	if err != nil {
		kv.ui.Error(fmt.Sprintf("not able to read config: %s", err.Error()))
		return 1
	}

	if config.RootToken == "" {
		kv.ui.Error("missing root token")
		return 1
	}

	rootTokenBytes, err := base64.StdEncoding.DecodeString(config.RootToken)

	if err != nil {
		kv.ui.Error(fmt.Sprintf("not able to decode root token: %s", err.Error()))
		return 1
	}

	path := fs.Arg(0)

	if path == "" {
		kv.ui.Error("missing path")
		return 1
	}

	if strings.HasPrefix(path, internal.KeyringPrefix) {
		kv.ui.Warn(fmt.Sprintf("paths prefixed with %s are restricted", internal.KeyringPrefix))
		return 1
	}

	kvPairs, err := internal.ParseArgsData(os.Stdin, fs.Args()[1:])

	if err != nil {
		kv.ui.Error(fmt.Sprintf("failed to parse key value pairs: %s", err.Error()))
		return 1
	}

	if len(kvPairs) == 0 {
		kv.ui.Error("must provide key value data")
		return 1
	}

	kvbackend, err := kv.setupBackend(defaultCtx, config)

	if err != nil {
		kv.ui.Error(fmt.Sprintf("failed to setup backend: %s", err.Error()))
		return 1
	}

	defer func() {
		cleanupErr := kvbackend.Cleanup(defaultCtx)
		if cleanupErr != nil {
			kv.ui.Warn(fmt.Sprintf("failed to gracefully shutdown backend: %s", cleanupErr.Error()))
		}
	}()

	barrier, err := internal.NewBarrier(kvbackend)

	if err != nil {
		kv.ui.Error(fmt.Sprintf("failed to instantiate barrier: %s", err.Error()))
		return 1
	}

	initialized, err := barrier.KeyringPersisted(defaultCtx)

	if err != nil {
		kv.ui.Error(fmt.Sprintf("failed to validate existing barrier: %s", err.Error()))
		return 1
	}

	if !initialized {
		kv.ui.Warn("keyring not initialized. Run keyring init.")
		return 1
	}

	err = barrier.Initialize(defaultCtx, string(rootTokenBytes))

	if err != nil {
		kv.ui.Error(fmt.Sprintf("failed to initialize existing barrier: %s", err.Error()))
		return 1
	}

	entries := make([]*backend.BackendEntry, 0)

	for key, pair := range kvPairs {

		entryBytes, err := json.Marshal(pair)

		if err != nil {
			kv.ui.Error(fmt.Sprintf("failed to convert key value into bytes: %s", err.Error()))
			return 1
		}

		entries = append(entries, &backend.BackendEntry{
			Key:   key,
			Value: entryBytes,
		})
	}

	err = barrier.Put(defaultCtx, path, entries)

	if err != nil {
		kv.ui.Error(fmt.Sprintf("failed to persist kv pairs at path '%s': %s", path, err.Error()))
		return 1
	}

	kv.ui.Info("success!")
	return 0
}

func (kv *KVPutCommand) readConfig(args []string) (*internal.GeneralConfig, *flag.FlagSet, error) {

	config := &internal.GeneralConfig{
		Backend: &internal.BackendConfig{},
	}

	fs := flag.NewFlagSet("put", flag.ContinueOnError)
	fs.StringVar(&config.RootToken, "root-token", "", "The root token to use for encryption")
	fs.StringVar(&config.Backend.Type, "backend-type", "", "The type of backend to use")

	// filebackend type
	fileConfig := &bbolt.Config{}
	fs.StringVar(&fileConfig.Path, "filepath", "", "The file path to the local datastore")

	if err := fs.Parse(args); err != nil {
		return nil, nil, err
	}

	config = internal.DefaultGeneralConfig().Merge(config)

	backendType, err := config.Backend.GetBackendType()

	if err != nil {
		return nil, nil, err
	}

	if backendType == backend.FileBackend {
		config.Backend.Options = bbolt.DefaultConfig().Merge(fileConfig)
	} else {
		return nil, nil, fmt.Errorf("backend options not setup for backend type '%s'", backendType)
	}

	return config, fs, nil
}

func (kv *KVPutCommand) setupBackend(ctx context.Context, config *internal.GeneralConfig) (backend.Backend, error) {

	backendType, err := config.Backend.GetBackendType()

	if err != nil {
		return nil, err
	}

	if backendType == backend.FileBackend {

		backendConfig, ok := config.Backend.Options.(*bbolt.Config)

		if !ok {
			return nil, fmt.Errorf("failed to cast backend config to file backend config")
		}

		fileBackend, err := bbolt.NewBoltBackend(backendConfig)

		if err != nil {
			return nil, err
		}

		err = fileBackend.Setup(ctx)

		if err != nil {
			return nil, err
		}

		return fileBackend, nil
	}

	return nil, nil
}
