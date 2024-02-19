package app

import (
	_ "embed" // embed is used to embed the env.json file
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

//nolint:golint
//go:embed env.json
var EnvJSON []byte

func SetEnvJSON(envJSON []byte) error {
	const envJSONPermission = os.FileMode(0600)

	err := os.WriteFile(filepath.Join(Root, "app", "env.json"), envJSON, envJSONPermission)
	if err != nil {
		return fmt.Errorf("failed to write env json: %w", err)
	}

	return nil
}

//nolint:gochecknoglobals
var (
	_, b, _, _ = runtime.Caller(0)
	Root       = filepath.Join(filepath.Dir(b), "..")
)
