package util

import (
	"archive/zip"
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

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
	Runnables []RawWASM
}

// RawWASM is the raw data from a WASM file
type RawWASM struct {
	Name     string
	Contents []byte
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

	bundle := &Bundle{make([]RawWASM, len(r.File))}

	// Iterate through the files in the archive,

	for i, f := range r.File {
		raw := RawWASM{
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

		raw.Contents = bytes

		bundle.Runnables[i] = raw
	}

	return bundle, nil
}
