package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/suborbital/hive-wasm/bundle"
	"github.com/suborbital/hive-wasm/directive"
)

func main() {
	files := []os.File{}
	for _, filename := range []string{"helloworld-rs.wasm", "hivew_rs_builder.wasm", "swiftc_runnable.wasm"} {
		path := filepath.Join("./", "wasm", "testdata", filename)

		file, err := os.Open(path)
		if err != nil {
			log.Fatal("failed to open file", err)
		}

		files = append(files, *file)
	}

	directive := &directive.Directive{
		Identifier: "dev.suborbital.appname",
		Version:    "v0.1.1",
		Functions: []directive.Function{
			{
				Name:      "helloworld-rs",
				NameSpace: "default",
			},
			{
				Name:      "hivew_rs_builder",
				NameSpace: "default",
			},
			{
				Name:      "swiftc_runnable",
				NameSpace: "default",
			},
		},
		Handlers: []directive.Handler{
			{
				Input: directive.Input{
					Type:     directive.InputTypeRequest,
					Method:   "GET",
					Resource: "/api/v1/user",
				},
				Steps: []directive.Executable{
					{
						Fn: "swiftc_runnable",
					},
					{
						Fn: "helloworld-rs",
					},
					{
						Fn: "hivew_rs_builder",
					},
				},
			},
		},
	}

	if err := directive.Validate(); err != nil {
		log.Fatal("failed to validate directive", err)
	}

	if err := bundle.Write(directive, files, "./runnables.wasm.zip"); err != nil {
		log.Fatal("failed to WriteBundle", err)
	}

	fmt.Println("done ✨")
}