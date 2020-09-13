package wasm

import (
	"sync"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/wasmerio/wasmer-go/wasmer"
)

// #include <stdlib.h>
//
// extern void return_result(void *context, int32_t pointer, int32_t size, int32_t envIndex, int32_t instIndex);
import "C"

//export return_result
func return_result(context unsafe.Pointer, pointer int32, size int32, envIndex int32, instIndex int32) {
	envLock.RLock()
	defer envLock.RUnlock()

	env := environments[envIndex]
	inst := env.instances[instIndex]

	memory := inst.wasmerInst.Memory.Data()[pointer:]

	result := make([]byte, size)

	for index := 0; int32(index) < size; index++ {
		result[index] = memory[index]
	}

	inst.resultChan <- result
}

var environments []*wasmEnvironment
var envOnce = sync.Once{}
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

	// meta related to this env's position in the shared array, and the index of the last used wasmInstance
	envIndex  int
	instIndex int
}

// newEnvironment creates a new environment and adds it to the shared environments array
// such that WASM instances can return data to the correct place
func newEnvironment(name string, filepath string) *wasmEnvironment {
	envOnce.Do(func() {
		environments = []*wasmEnvironment{}
	})

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

// setRaw sets the raw bytes of a WASM module to be used rather than a filepath
func (w *wasmEnvironment) setRaw(raw []byte) {
	w.raw = raw
}

func (w *wasmInstance) writeInput(input []byte) int32 {
	lengthOfInput := len(input)

	// Allocate memory for the input, and get a pointer to it.
	allocateResult, _ := w.wasmerInst.Exports["allocate_input"](lengthOfInput)
	inputPointer := allocateResult.ToI32()

	// Write the input into the memory.
	memory := w.wasmerInst.Memory.Data()[inputPointer:]

	for index := 0; index < lengthOfInput; index++ {
		memory[index] = input[index]
	}

	return inputPointer
}

func (w *wasmInstance) deallocate(pointer int32, length int) {
	dealloc := w.wasmerInst.Exports["deallocate"]

	dealloc(pointer, length)
}
