package wasm

import (
	"sync"

	"github.com/pkg/errors"
	wasm "github.com/wasmerio/go-ext-wasm/wasmer"
)

// Environment is the raw data from a WASM file
type Environment struct {
	Name         string
	wasmFilePath string
	Raw          []byte
	instances    []instance
	index        int
}

type instance struct {
	instance wasm.Instance
	lock     sync.Mutex
}

// getInstance returns a wasmer instance
func (c *Environment) useInstance(instFunc func(wasm.Instance)) {
	if c.index == len(c.instances)-1 {
		c.index = 0
	} else {
		c.index++
	}

	inst := &c.instances[c.index]
	inst.lock.Lock()
	defer inst.lock.Unlock()

	instFunc(inst.instance)
}

func (c *Environment) addInstance() error {
	if c.instances == nil {
		c.instances = []instance{}
		c.index = 0
	}

	if c.Raw == nil || len(c.Raw) == 0 {
		bytes, err := wasm.ReadBytes(c.wasmFilePath)
		if err != nil {
			return errors.Wrap(err, "failed to ReadBytes")
		}

		c.Raw = bytes
	}

	imports, err := wasm.NewDefaultWasiImportObjectForVersion(wasm.Snapshot1).Imports()
	if err != nil {
		return errors.Wrap(err, "failed to create Imports")
	}

	inst, err := wasm.NewInstanceWithImports(c.Raw, imports)
	if err != nil {
		return errors.Wrap(err, "failed to NewInstance")
	}

	instance := instance{
		instance: inst,
		lock:     sync.Mutex{},
	}

	c.instances = append(c.instances, instance)

	return nil
}
