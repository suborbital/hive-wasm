package wasm

import (
	"encoding/json"

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
	inputBytes, err := interfaceToBytes(job.Data())
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert job data to bytes for WASM Runnable")
	}

	var output []byte

	w.env.useInstance(func(instance *wasmInstance) {
		inPointer := instance.writeInput(inputBytes)

		wasmRun := instance.wasmerInst.Exports["run_e"]

		if _, wasmErr := wasmRun(inPointer, len(inputBytes), instance.envIndex, instance.instIndex); wasmErr != nil {
			err = errors.Wrap(wasmErr, "failed to wasmRun")
			return
		}

		output = <-instance.resultChan

		// deallocate the memory used for the input
		instance.deallocate(inPointer, len(inputBytes))
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

func interfaceToBytes(data interface{}) ([]byte, error) {
	// if data is []byte or string, return it as-is
	if b, ok := data.([]byte); ok {
		return b, nil
	} else if s, ok := data.(string); ok {
		return []byte(s), nil
	}

	// otherwise, assume it's a struct of some kind,
	// so JSON marshal it and return it
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to Marshal job data")
	}

	return dataJSON, nil
}
