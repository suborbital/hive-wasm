package wasm

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/suborbital/hive/hive"
)

// HandleBundle loads a .wasm.zip file into the hive instance
func HandleBundle(h *hive.Hive, path string) error {
	if !strings.HasSuffix(path, ".wasm.zip") {
		return fmt.Errorf("cannot load bundle %s, does not have .wasm.zip extension", filepath.Base(path))
	}

	bundle, err := ReadBundle(path)
	if err != nil {
		return errors.Wrap(err, "failed to ReadBundle")
	}

	for i, r := range bundle.Runnables {
		runner := newRunnerFromEnvironment(&bundle.Runnables[i])

		jobName := strings.Replace(r.Name, ".wasm", "", -1)
		h.Handle(jobName, runner)
	}

	return nil
}

// based loosely on https://golang.org/src/archive/zip/example_test.go

// WriteBundle writes a runnable bundle
func WriteBundle(files []os.File, targetPath string) error {
	// Create a buffer to write our archive to.
	buf := new(bytes.Buffer)

	// Create a new zip archive.
	w := zip.NewWriter(buf)

	// Add some files to the archive.

	for _, file := range files {
		f, err := w.Create(filepath.Base(file.Name()))
		if err != nil {
			return errors.Wrapf(err, "failed to add %s to bundle", file.Name())
		}

		contents, err := ioutil.ReadAll(&file)
		if err != nil {
			return errors.Wrapf(err, "failed to read file %s", file.Name())
		}

		_, err = f.Write(contents)
		if err != nil {
			return errors.Wrapf(err, "failed to write %s into bundle", file.Name())
		}

	}

	if err := w.Close(); err != nil {
		return errors.Wrap(err, "failed to close bundle writer")
	}

	if err := ioutil.WriteFile(targetPath, buf.Bytes(), 0700); err != nil {
		return errors.Wrap(err, "failed to write bundle to disk")
	}

	return nil
}

// Bundle represents a Runnable bundle
type Bundle struct {
	Runnables []Environment
}

// ReadBundle reads a .wasm.zip file and returns the bundle of wasm files within as raw bytes
// (suitable to be loaded into a wasmer instance)
func ReadBundle(path string) (*Bundle, error) {
	// Open a zip archive for reading.
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open bundle")
	}

	defer r.Close()

	bundle := &Bundle{make([]Environment, len(r.File))}

	// Iterate through the files in the archive,

	for i, f := range r.File {
		ctx := Environment{
			Name: f.Name,
		}

		rc, err := f.Open()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to open %s from bundle", f.Name)
		}

		bytes, err := ioutil.ReadAll(rc)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read %s from bundle", f.Name)
		}

		rc.Close()

		ctx.Raw = bytes

		bundle.Runnables[i] = ctx
	}

	return bundle, nil
}
