package wasm

// #include <stdlib.h>
//
// extern void return_result(void *context, int32_t pointer, int32_t size, int32_t envIndex, int32_t instIndex);
import "C"
import (
	"net/http"
	"unsafe"
)

/*
 In order to allow "easy" communication of data across the FFI barrier (outbound Go -> WASM and inbound WASM -> Go), hivew provides
 an FFI API. Functions exported from a WASM module can be easily called by Go code via the Wasmer instance exports, but returning data
 to the host Go code is not quite as straightforward.

 In order to accomplish this, hivew internally keeps a set of "environments" in a singleton package var (`environments` below).
 Each environment is a container that includes the WASM module bytes, and a set of WASM instances (runtimes) to execute said module.
 The envionment object has an index referencing its place in the singleton array, and each instance has an index referencing its position within
 the environment's instance array.

 When a WASM function calls one of the FFI API functions, it includes the `env_index` and `inst_index` values that were provided at the beginning
 of job execution, which allows hivew to look up the [env][instance] and send the result on the appropriate result channel. This is needed due to
 the way Go makes functions available on the FFI using CGO.
*/

///////////////////////////////////////////////////////////////
// below is the "hivew API" used for cross-FFI communication //
///////////////////////////////////////////////////////////////

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

// export fetch
func fetch(context unsafe.Pointer, urlPointer int32, urlSize int32, envIndex int32, instIndex int32) int32 {
	// fetch makes a network request on bahalf of the wasm runner.
	// fetch writes the http response body into memory starting at returnBodyPointer, and the return value is a pointer to that memory
	urlBytes, err := 
	req := http.NewRequest(http.MethodGet, )
	return 0
}

/////////////////////////////////////////////////////////////////////////////
// below is the wasm glue code used to manipulate wasm instance memory     //
// this requires a set of functions to be available within the wasm module //
// - allocate_input                                                        //
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
