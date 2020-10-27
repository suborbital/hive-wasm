package directive

import "gopkg.in/yaml.v2"

// Directive describes a set of functions and a set of handlers
// that take an input, and compose a set of functions to handle it
type Directive struct {
	Version   string
	Functions []Function
	Handlers  []Handler
}

// Marshal outputs the YAML bytes of the Directive
func (d *Directive) Marshal() ([]byte, error) {
	return yaml.Marshal(d)
}

// Unmarshal unmarshals YAML bytes into a Directive struct
func (d *Directive) Unmarshal(in []byte) error {
	return yaml.Unmarshal(in, d)
}

// Function describes a function present inside of a bundle
type Function struct {
	Name      string
	NameSpace string
}

// Handler represents the mapping between an input and a composition of functions
type Handler struct {
	Input Input `yaml:"input,inline"`
	Steps []Executable
}

// Input represents an input source
type Input struct {
	Type string
	// Some kind of metadata here?
}

// Executable represents an executable step in a handler
type Executable interface {
	Type() string
}

// Group represents a group of functions
type Group struct {
	Group []Single
}

// Type returns the executable type
func (g Group) Type() string { return "group" }

// Single represents a singe function
type Single struct {
	FQFN string // FQFN stands for "fully qualified function name" in the style namespace#functionname@version.number.semver
}

// Type returns the executable type
func (s Single) Type() string { return "single" }
