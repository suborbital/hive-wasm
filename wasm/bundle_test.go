package wasm

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"
)

func TestReadBundle(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Error(errors.Wrap(err, "failed to get CWD"))
	}

	bundle, err := ReadBundle(filepath.Join(cwd, "./testdata/runnables.wasm.zip"))
	if err != nil {
		t.Error(errors.Wrap(err, "failed to ReadBundle"))
		return
	}

	if len(bundle.Runnables) == 0 {
		t.Error("bundle had 0 runnables")
		return
	}

	hasDefault := false
	for _, r := range bundle.Runnables {
		if r.Name == "helloworld-rs.wasm" {
			hasDefault = true
		}
	}

	if !hasDefault {
		t.Error("default helloworld-rs.wasm not found in bundle")
	}
}
