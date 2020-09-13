package wasm

import (
	"github.com/suborbital/hive/hive"

	"github.com/pkg/errors"
)

//Runner represents a wasm-based runnable
type Runner struct {
	env *wasmEnvironment
}

// NewRunner returns a new *Runner
func NewRunner(filepath string) *Runner {
	return newRunnerWithEnvironment(newEnvironment("", filepath))
}

func newRunnerWithEnvironment(env *wasmEnvironment) *Runner {
	w := &Runner{
		env: env,
	}

	return w
}

// Run runs a Runner
func (w *Runner) Run(job hive.Job, do hive.DoFunc) (interface{}, error) {
	input := job.Bytes()
	if input == nil {
		return nil, errors.New("WASM jobs must be []byte")
	}

	var output []byte
	var err error

	w.env.useInstance(func(instance *wasmInstance) {
		inPointer := instance.writeInput(input)

		wasmRun := instance.wasmerInst.Exports["run_e"]

		if _, wasmErr := wasmRun(inPointer, len(input), instance.envIndex, instance.instIndex); wasmErr != nil {
			err = errors.Wrap(wasmErr, "failed to wasmRun")
			return
		}

		output = <-instance.resultChan

		// deallocate the memory used for the input
		instance.deallocate(inPointer, len(input))
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
