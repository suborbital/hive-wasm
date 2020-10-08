package wasm

// #include <stdlib.h>
//
// extern void return_result(void *context, int32_t pointer, int32_t size, int32_t envIndex, int32_t instIndex);
// extern int32_t fetch(void *context, int32_t urlPointer, int32_t urlSize, int32_t destPointer, int32_t destMaxSize, int32_t envIndex, int32_t instIndex);
// extern void print(void *context, int32_t pointer, int32_t size, int32_t envIndex, int32_t instIndex);
import "C"

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"unsafe"

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
	imports.AppendFunction("fetch", fetch, C.fetch)
	imports.AppendFunction("print", print, C.print)

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

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// below is the "hivew API" which grants capabilites to WASM runnables by routing things like network requests through the host (Go) code //
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

//export return_result
func return_result(context unsafe.Pointer, pointer int32, size int32, envIndex int32, instIndex int32) {
	// TODO: make it impossible for a module to call out to another instance (obfucate the indices?)
	envLock.RLock()
	defer envLock.RUnlock()

	inst := instanceAtIndices(envIndex, instIndex)
	if inst == nil {
		// not sure what to do here
		return
	}

	result := inst.readMemory(pointer, size)

	inst.resultChan <- result
}

//export fetch
func fetch(context unsafe.Pointer, urlPointer int32, urlSize int32, destPointer int32, destMaxSize int32, envIndex int32, instIndex int32) int32 {
	// fetch makes a network request on bahalf of the wasm runner.
	// fetch writes the http response body into memory starting at returnBodyPointer, and the return value is a pointer to that memory
	inst := instanceAtIndices(envIndex, instIndex)
	if inst == nil {
		fmt.Println("couldn't find inst")
		return -1
	}

	urlBytes := inst.readMemory(urlPointer, urlSize)

	urlObj, err := url.Parse(string(urlBytes))
	if err != nil {
		fmt.Println("couldn't parse URL")
		return -2
	}

	req, err := http.NewRequest(http.MethodGet, urlObj.String(), nil)
	if err != nil {
		fmt.Println("failed to build request")
		return -2
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("failed to Do request")
		return -3
	}

	defer resp.Body.Close()
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("failed to Read response body")
		return -4
	}

	if len(respBytes) <= int(destMaxSize) {
		inst.writeMemoryAtLocation(destPointer, respBytes)
	}

	return int32(len(respBytes))
}

//export print
func print(context unsafe.Pointer, pointer int32, size int32, envIndex int32, instIndex int32) {
	inst := instanceAtIndices(envIndex, instIndex)
	if inst == nil {
		fmt.Println("print: couldn't find inst")
	}

	msgBytes := inst.readMemory(pointer, size)
	msg := fmt.Sprintf("[%d:%d]: %s", envIndex, instIndex, string(msgBytes))

	fmt.Println(msg)
}
