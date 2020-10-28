package directive

import (
	"testing"
)

func TestYAMLMarshalUnmarshal(t *testing.T) {
	dir := Directive{
		Version: "v0.1.1",
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
					Type:     "request",
					Resource: "/api/v1/user",
				},
				Steps: []Executable{
					{
						Group: []string{
							"db#getUser@0.1.1",
							"db#getUserDetails@0.1.1",
						},
					},
					{
						Fn: "db#returnUser@0.1.1",
					},
				},
			},
		},
	}

	yamlBytes, err := dir.Marshal()
	if err != nil {
		t.Error(err)
		return
	}

	dir2 := Directive{}
	if err := dir2.Unmarshal(yamlBytes); err != nil {
		t.Error(err)
		return
	}

	if err := dir2.Validate(); err != nil {
		t.Error(err)
	}

	if len(dir2.Handlers[0].Steps) != 2 {
		t.Error("wrong number of steps")
		return
	}

	if len(dir2.Functions) != 3 {
		t.Error("wrong number of steps")
		return
	}
}
