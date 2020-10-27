package directive

import (
	"fmt"
	"testing"
)

func TestYAMLMarshalUnmarshal(t *testing.T) {
	dir := Directive{
		Version: "0.1.1",
		Functions: []Function{
			{
				Name:      "getUser",
				NameSpace: "db",
			},
			{
				Name:      "getUserDetails",
				NameSpace: "db",
			},
			{
				Name:      "returnUser",
				NameSpace: "api",
			},
		},
		Handlers: []Handler{
			{
				Input: Input{
					Type: "request",
				},
				Steps: []Executable{
					Group{Group: []Single{
						{FQFN: "db#getUser@0.1.1"},
						{FQFN: "db#getUserDetails@0.1.1"},
					}},
					Single{FQFN: "db#returnUser@0.1.1"},
				},
			},
		},
	}

	yamlBytes, err := dir.Marshal()
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(string(yamlBytes))
}
