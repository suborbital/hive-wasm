package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/suborbital/hive-wasm/directive"
	"github.com/suborbital/hive-wasm/wasm"
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
					Type:     "request",
					Method:   "GET",
					Resource: "/api/v1/user",
				},
				Steps: []directive.Executable{
					// {
					// 	Group: []string{
					// 		"db#getUser@0.1.1",
					// 		"db#getUserDetails@0.1.1",
					// 	},
					// },
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

	if err := wasm.WriteBundle(directive, files, "./runnables.wasm.zip"); err != nil {
		log.Fatal("failed to WriteBundle", err)
	}

	fmt.Println("done ✨")
}
