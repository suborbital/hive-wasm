package wasm

import (
	"strings"

	"github.com/suborbital/hive/hive"

	"github.com/pkg/errors"

	wasm "github.com/wasmerio/go-ext-wasm/wasmer"
)

//Runner represents a wasm-based runnable
type Runner struct {
	env *Environment
}

// NewRunner returns a new *Runner
func NewRunner(path string) *Runner {
	w := &Runner{
		env: &Environment{
			wasmFilePath: path,
		},
	}

	return w
}

func newRunnerFromEnvironment(env *Environment) *Runner {
	w := &Runner{
		env: env,
	}

	return w
}

// Run runs a Runner
func (w *Runner) Run(job hive.Job, do hive.DoFunc) (interface{}, error) {
	input, ok := job.Data().(string)
	if !ok {
		return nil, errors.New("failed to run WASM job, input is not string")
	}

	var output string
	var err error

	w.env.useInstance(func(instance wasm.Instance) {
		inPointer := writeInput(instance, input)

		wasmRun := instance.Exports["run_e"]

		res, err := wasmRun(inPointer)
		if err != nil {
			err = errors.Wrap(err, "failed to wasmRun")
		}

		output = readOutput(instance, res.ToI32())

		// deallocate the memory used for the input and output
		deallocate(instance, inPointer, len(input))
		deallocate(instance, res.ToI32(), len(output))
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

// OnStart runs when a worker starts using this Runnable
func (w *Runner) OnStart() error {
	if err := w.env.addInstance(); err != nil {
		return errors.Wrap(err, "failed to addInstance")
	}

	return nil
}

func writeInput(inst wasm.Instance, input string) int32 {
	lengthOfInput := len(input)

	// Allocate memory for the input, and get a pointer to it.
	allocateResult, _ := inst.Exports["allocate_input"](lengthOfInput)
	inputPointer := allocateResult.ToI32()

	// Write the input into the memory.
	memory := inst.Memory.Data()[inputPointer:]

	for nth := 0; nth < lengthOfInput; nth++ {
		memory[nth] = input[nth]
	}

	// C-string terminates by NULL.
	memory[lengthOfInput] = 0

	return inputPointer
}

func readOutput(inst wasm.Instance, pointer int32) string {
	memory := inst.Memory.Data()[pointer:]

	nth := 0
	var output strings.Builder

	for {
		if memory[nth] == 0 {
			break
		}

		output.WriteByte(memory[nth])
		nth++
	}

	return output.String()
}

func deallocate(inst wasm.Instance, pointer int32, length int) {
	dealloc := inst.Exports["deallocate"]

	dealloc(pointer, length)
}
