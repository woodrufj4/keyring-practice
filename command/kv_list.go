package command

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/mitchellh/cli"
	"github.com/woodrufj4/keyring-practice/internal"
)

type KVListCommand struct {
	ui cli.Ui
}

func (kv KVListCommand) Synopsis() string {
	return "Lists the stored paths given a path prefix"
}

func (kv KVListCommand) Help() string {
	helpText := `
Usage: keying list [options] <path-prefix>

  Lists the stored paths given a path prefix

  Example:

    $ keyring list secret

  Options:

    -root-token=<string>
      The root token to access the keyring.

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

func (kv *KVListCommand) Run(args []string) int {

	defaultCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	config, fs, err := internal.ReadConfig(args)

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
		kv.ui.Error("missing path prefix")
		return 1
	}

	if strings.HasPrefix(path, internal.KeyringPrefix) {
		kv.ui.Warn(fmt.Sprintf("paths prefixed with %s are restricted", internal.KeyringPrefix))
		return 1
	}

	kvbackend, err := internal.SetupBackend(defaultCtx, config)

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

	pathNames, err := barrier.List(defaultCtx, path)

	if err != nil {
		kv.ui.Error(fmt.Sprintf("failed to list paths at prefix %s", err.Error()))
		return 1
	}

	for _, pathName := range pathNames {
		kv.ui.Output(pathName)
	}

	return 0
}
