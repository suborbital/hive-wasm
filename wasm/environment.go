package wasm

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/wasmerio/wasmer-go/wasmer"
)

var environments []*wasmEnvironment
var envLock = sync.RWMutex{}

// wasmEnvironment is an wasmEnvironment in which WASM instances run
type wasmEnvironment struct {
	Name      string
	filepath  string
	raw       []byte
	instances []*wasmInstance

	// meta related to this env's position in the shared array, and the index of the last used wasmInstance
	envIndex  int
	instIndex int
}

type wasmInstance struct {
	wasmerInst wasmer.Instance
	resultChan chan []byte
	lock       sync.Mutex

	// meta related to this env's position in the shared array, and the index of this wasmInstance in the environment
	envIndex  int
	instIndex int
}

// newEnvironment creates a new environment and adds it to the shared environments array
// such that WASM instances can return data to the correct place
func newEnvironment(name string, filepath string) *wasmEnvironment {
	envLock.Lock()
	defer envLock.Unlock()

	e := &wasmEnvironment{
		Name:      name,
		filepath:  filepath,
		instances: []*wasmInstance{},
		envIndex:  len(environments),
		instIndex: 0,
	}

	environments = append(environments, e)

	return e
}

// useInstance provides an instance from the environment's pool to be used
func (w *wasmEnvironment) useInstance(instFunc func(*wasmInstance)) {
	if w.instIndex == len(w.instances)-1 {
		w.instIndex = 0
	} else {
		w.instIndex++
	}

	inst := w.instances[w.instIndex]
	inst.lock.Lock()
	defer inst.lock.Unlock()

	instFunc(inst)
}

// addInstance adds a new WASM instance to the environment's pool
func (w *wasmEnvironment) addInstance() error {
	if w.raw == nil || len(w.raw) == 0 {
		bytes, err := wasmer.ReadBytes(w.filepath)
		if err != nil {
			return errors.Wrap(err, "failed to ReadBytes")
		}

		w.raw = bytes
	}

	instance := &wasmInstance{
		resultChan: make(chan []byte, 1),
		lock:       sync.Mutex{},
		envIndex:   w.envIndex,
		instIndex:  len(w.instances),
	}

	imports, err := wasmer.NewDefaultWasiImportObjectForVersion(wasmer.Snapshot1).Imports()
	if err != nil {
		return errors.Wrap(err, "failed to create Imports")
	}

	imports.AppendFunction("return_result", return_result, C.return_result)

	inst, err := wasmer.NewInstanceWithImports(w.raw, imports)
	if err != nil {
		return errors.Wrap(err, "failed to NewInstance")
	}

	instance.wasmerInst = inst

	w.instances = append(w.instances, instance)

	return nil
}

func instanceAtIndices(envIndex int32, instIndex int32) *wasmInstance {
	if int(envIndex) > len(environments)-1 {
		return nil
	}

	env := environments[envIndex]

	if int(instIndex) > len(env.instances)-1 {
		return nil
	}

	return env.instances[instIndex]
}

// setRaw sets the raw bytes of a WASM module to be used rather than a filepath
func (w *wasmEnvironment) setRaw(raw []byte) {
	w.raw = raw
}

func init() {
	environments = []*wasmEnvironment{}
	envLock = sync.RWMutex{}
}
