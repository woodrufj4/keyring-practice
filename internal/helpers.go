package internal

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"

	kvbuilder "github.com/hashicorp/go-secure-stdlib/kv-builder"
	"github.com/ryanuber/columnize"
	"github.com/woodrufj4/keyring-practice/backend"
	"github.com/woodrufj4/keyring-practice/backend/bbolt"
)

// parseArgsData parses the given args in the format key=value into a map of
// the provided arguments. The given reader can also supply key=value pairs.
func ParseArgsData(stdin io.Reader, args []string) (map[string]interface{}, error) {

	builder := &kvbuilder.Builder{Stdin: stdin}

	if err := builder.Add(args...); err != nil {
		return nil, err
	}

	return builder.Map(), nil

}

func ReadConfig(args []string) (*GeneralConfig, *flag.FlagSet, error) {

	config := &GeneralConfig{
		Backend: &BackendConfig{},
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

	config = DefaultGeneralConfig().Merge(config)

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

func SetupBackend(ctx context.Context, config *GeneralConfig) (backend.Backend, error) {

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

func FormatTableOutput(entries []*backend.BackendEntry) (string, error) {

	outputConfig := columnize.Config{
		Delim: "♨",
		Glue:  "    ",
		Empty: "n/a",
	}

	outputArray := make([]string, 0)

	outputArray = append(outputArray, fmt.Sprintf("Key%sValue", outputConfig.Delim))
	outputArray = append(outputArray, fmt.Sprintf("---%s-----", outputConfig.Delim))

	var entryValue interface{}

	for _, entry := range entries {

		if err := json.Unmarshal(entry.Value, &entryValue); err != nil {
			return "", err
		}

		outputArray = append(outputArray, fmt.Sprintf("%s%s%v", entry.Key, outputConfig.Delim, entryValue))
	}

	return columnize.Format(outputArray, &outputConfig), nil
}
