package wasm

// #include <stdlib.h>
//
// extern void return_result(void *context, int32_t pointer, int32_t size, int32_t ident);
// extern void return_result_swift(void *context, int32_t pointer, int32_t size, int32_t ident, int32_t swiftself, int32_t swifterr);
//
// extern int32_t fetch_url(void *context, int32_t urlPointer, int32_t urlSize, int32_t destPointer, int32_t destMaxSize, int32_t ident);
//
// extern void log_msg(void *context, int32_t pointer, int32_t size, int32_t level, int32_t ident);
// extern void log_msg_swift(void *context, int32_t pointer, int32_t size, int32_t level, int32_t ident, int32_t swiftself, int32_t swifterr);
import "C"

import (
	"crypto/rand"
	"math"
	"math/big"
	"sync"

	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"unsafe"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/wasmerio/wasmer-go/wasmer"
)

/*
 In order to allow "easy" communication of data across the FFI barrier (outbound Go -> WASM and inbound WASM -> Go), hivew provides
 an FFI API. Functions exported from a WASM module can be easily called by Go code via the Wasmer instance exports, but returning data
 to the host Go code is not quite as straightforward.

 In order to accomplish this, hivew internally keeps a set of "environments" in a singleton package var (`environments` below).
 Each environment is a container that includes the WASM module bytes, and a set of WASM instances (runtimes) to execute said module.
 The envionment object has an index referencing its place in the singleton array, and each instance has an index referencing its position within
 the environment's instance array.

 When a WASM function calls one of the FFI API functions, it includes the `ident`` value that was provided at the beginning
 of job execution, which allows hivew to look up the [env][instance] and send the result on the appropriate result channel. This is needed due to
 the way Go makes functions available on the FFI using CGO.
*/

// the globally shared set of Wasm environments, accessed by UUID
var environments = map[string]*wasmEnvironment{}

// a lock to ensure the environments array is concurrency safe (didn't use sync.Map to prevent type coersion)
var envLock = sync.RWMutex{}

// the instance mapper maps a random int32 to a wasm instance to prevent malicious access to other instances via the FFI
var instanceMapper = sync.Map{}

// wasmEnvironment is an environmenr in which Wasm instances run
type wasmEnvironment struct {
	Name      string
	UUID      string
	filepath  string
	raw       []byte
	instances []*wasmInstance

	// the index of the last used wasm instance
	instIndex int
	lock      sync.Mutex
}

type wasmInstance struct {
	wasmerInst wasmer.Instance
	resultChan chan []byte
	lock       sync.Mutex
}

// instanceReference is a "pointer" to the global environments array and the
// wasm instances within each environment
type instanceReference struct {
	EnvUUID   string
	InstIndex int
}

// newEnvironment creates a new environment and adds it to the shared environments array
// such that Wasm instances can return data to the correct place
func newEnvironment(name string, filepath string) *wasmEnvironment {
	envLock.Lock()
	defer envLock.Unlock()

	e := &wasmEnvironment{
		Name:      name,
		UUID:      uuid.New().String(),
		filepath:  filepath,
		instances: []*wasmInstance{},
		instIndex: 0,
		lock:      sync.Mutex{},
	}

	environments[e.UUID] = e

	return e
}

// useInstance provides an instance from the environment's pool to be used
func (w *wasmEnvironment) useInstance(instFunc func(*wasmInstance, int32)) error {
	w.lock.Lock()

	if w.instIndex == len(w.instances)-1 {
		w.instIndex = 0
	} else {
		w.instIndex++
	}

	instIndex := w.instIndex
	inst := w.instances[instIndex]

	w.lock.Unlock() // now that we've acquired our instance, let the next one go

	inst.lock.Lock()
	defer inst.lock.Unlock()

	// generate a random identifier as a reference to the instance in use to
	// easily allow the Wasm module to reference itself when calling back over the FFI
	ident, err := setupNewIdentifier(w.UUID, instIndex)
	if err != nil {
		return errors.Wrap(err, "failed to setupNewIdentifier")
	}

	instFunc(inst, ident)

	removeIdentifier(ident)

	return nil
}

// addInstance adds a new Wasm instance to the environment's pool
func (w *wasmEnvironment) addInstance() error {
	w.lock.Lock()
	defer w.lock.Unlock()

	if w.raw == nil || len(w.raw) == 0 {
		bytes, err := wasmer.ReadBytes(w.filepath)
		if err != nil {
			return errors.Wrap(err, "failed to ReadBytes")
		}

		w.raw = bytes
	}

	// mount the WASI interface
	imports, err := wasmer.NewDefaultWasiImportObjectForVersion(wasmer.Snapshot1).Imports()
	if err != nil {
		return errors.Wrap(err, "failed to create Imports")
	}

	// Mount the Runnable API
	imports.AppendFunction("return_result", return_result, C.return_result)
	imports.AppendFunction("return_result_swift", return_result_swift, C.return_result_swift)
	imports.AppendFunction("fetch_url", fetch_url, C.fetch_url)
	imports.AppendFunction("log_msg", log_msg, C.log_msg)
	imports.AppendFunction("log_msg_swift", log_msg_swift, C.log_msg_swift)

	inst, err := wasmer.NewInstanceWithImports(w.raw, imports)
	if err != nil {
		return errors.Wrap(err, "failed to NewInstance")
	}

	// if the module has exported an init, call it
	init := inst.Exports["init"]
	if init != nil {
		if _, err := init(); err != nil {
			return errors.Wrap(err, "failed to init instance")
		}
	}

	instance := &wasmInstance{
		wasmerInst: inst,
		resultChan: make(chan []byte, 1),
		lock:       sync.Mutex{},
	}

	w.instances = append(w.instances, instance)

	return nil
}

// setRaw sets the raw bytes of a Wasm module to be used rather than a filepath
func (w *wasmEnvironment) setRaw(raw []byte) {
	w.raw = raw
}

func setupNewIdentifier(envUUID string, instIndex int) (int32, error) {
	for {
		ident, err := randomIdentifier()
		if err != nil {
			return -1, errors.Wrap(err, "failed to randomIdentifier")
		}

		// ensure we don't accidentally overwrite something else
		// (however unlikely that may be)
		if _, exists := instanceMapper.Load(ident); exists {
			continue
		}

		ref := instanceReference{
			EnvUUID:   envUUID,
			InstIndex: instIndex,
		}

		instanceMapper.Store(ident, ref)

		return ident, nil
	}
}

func removeIdentifier(ident int32) {
	instanceMapper.Delete(ident)
}

func instanceForIdentifier(ident int32) (*wasmInstance, error) {
	rawRef, exists := instanceMapper.Load(ident)
	if !exists {
		return nil, errors.New("instance does not exist")
	}

	ref := rawRef.(instanceReference)

	envLock.RLock()
	defer envLock.RUnlock()

	env, exists := environments[ref.EnvUUID]
	if !exists {
		return nil, errors.New("environment does not exist")
	}

	if len(env.instances) <= ref.InstIndex-1 {
		return nil, errors.New("invalid instance index")
	}

	inst := env.instances[ref.InstIndex]

	return inst, nil
}

func randomIdentifier() (int32, error) {
	// generate a random number between 0 and the largest possible int32
	num, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt32))
	if err != nil {
		return -1, errors.Wrap(err, "failed to rand.Int")
	}

	return int32(num.Int64()), nil
}

/////////////////////////////////////////////////////////////////////////////
// below is the wasm glue code used to manipulate wasm instance memory     //
// this requires a set of functions to be available within the wasm module //
// - allocate                                                              //
// - deallocate                                                            //
/////////////////////////////////////////////////////////////////////////////

func (w *wasmInstance) readMemory(pointer int32, size int32) []byte {
	data := w.wasmerInst.Memory.Data()[pointer:]
	result := make([]byte, size)

	for index := 0; int32(index) < size; index++ {
		result[index] = data[index]
	}

	return result
}

func (w *wasmInstance) writeMemory(data []byte) (int32, error) {
	lengthOfInput := len(data)

	allocate := w.wasmerInst.Exports["allocate"]
	if allocate == nil {
		return -1, errors.New("missing required FFI function: allocate")
	}

	// Allocate memory for the input, and get a pointer to it.
	allocateResult, err := allocate(lengthOfInput)
	if err != nil {
		return -1, errors.Wrap(err, "failed to call allocate")
	}

	pointer := allocateResult.ToI32()

	w.writeMemoryAtLocation(pointer, data)

	return pointer, nil
}

func (w *wasmInstance) writeMemoryAtLocation(pointer int32, data []byte) {
	lengthOfInput := len(data)

	// Write the input into the memory.
	memory := w.wasmerInst.Memory.Data()[pointer:]

	for index := 0; index < lengthOfInput; index++ {
		memory[index] = data[index]
	}
}

func (w *wasmInstance) deallocate(pointer int32, length int) {
	dealloc := w.wasmerInst.Exports["deallocate"]

	dealloc(pointer, length)
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// below is the "Runnable API" which grants capabilites to Wasm runnables by routing things like network requests through the host (Go) code //
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

//export return_result
func return_result(context unsafe.Pointer, pointer int32, size int32, identifier int32) {
	envLock.RLock()
	defer envLock.RUnlock()

	inst, err := instanceForIdentifier(identifier)
	if err != nil {
		fmt.Println(errors.Wrap(err, "[hive-wasm] alert: invalid identifier used, potential malicious activity"))
		return
	}

	result := inst.readMemory(pointer, size)

	inst.resultChan <- result
}

//export return_result_swift
func return_result_swift(context unsafe.Pointer, pointer int32, size int32, identifier int32, swiftself int32, swifterr int32) {
	return_result(context, pointer, size, identifier)
}

//export fetch_url
func fetch_url(context unsafe.Pointer, urlPointer int32, urlSize int32, destPointer int32, destMaxSize int32, identifier int32) int32 {
	// fetch makes a network request on bahalf of the wasm runner.
	// fetch writes the http response body into memory starting at returnBodyPointer, and the return value is a pointer to that memory
	inst, err := instanceForIdentifier(identifier)
	if err != nil {
		fmt.Println(errors.Wrap(err, "[hive-wasm] alert: invalid identifier used, potential malicious activity"))
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

//export log_msg
func log_msg(context unsafe.Pointer, pointer int32, size int32, level int32, identifier int32) {
	inst, err := instanceForIdentifier(identifier)
	if err != nil {
		fmt.Println(errors.Wrap(err, "[hive-wasm] alert: invalid identifier used, potential malicious activity"))
		return
	}

	msgBytes := inst.readMemory(pointer, size)
	msg := fmt.Sprintf("[%d]: %s", identifier, string(msgBytes))

	fmt.Println(msg)
}

//export log_msg_swift
func log_msg_swift(context unsafe.Pointer, pointer int32, size int32, level int32, identifier int32, swiftself int32, swifterr int32) {
	log_msg(context, pointer, size, level, identifier)
}
