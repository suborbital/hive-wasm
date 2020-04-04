package util

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// derived loosely from https://golang.org/src/archive/zip/example_test.go

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

func ExampleReader() {

	// Open a zip archive for reading.

	r, err := zip.OpenReader("testdata/readme.zip")

	if err != nil {

		log.Fatal(err)

	}

	defer r.Close()

	// Iterate through the files in the archive,

	// printing some of their contents.

	for _, f := range r.File {

		fmt.Printf("Contents of %s:\n", f.Name)

		rc, err := f.Open()

		if err != nil {

			log.Fatal(err)

		}

		_, err = io.CopyN(os.Stdout, rc, 68)

		if err != nil {

			log.Fatal(err)

		}

		rc.Close()

		fmt.Println()

	}

}
