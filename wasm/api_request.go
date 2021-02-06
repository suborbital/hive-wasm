package wasm

import (
	"github.com/pkg/errors"
	"github.com/wasmerio/wasmer-go/wasmer"
)

const (
	fieldTypeMeta   = int32(0)
	fieldTypeBody   = int32(1)
	fieldTypeHeader = int32(2)
	fieldTypeParams = int32(3)
	fieldTypeState  = int32(4)
)

func requestGetField() *HostFn {
	fn := func(args ...wasmer.Value) (interface{}, error) {
		fieldType := args[0].I32()
		keyPointer := args[1].I32()
		keySize := args[2].I32()
		destPointer := args[3].I32()
		destMaxSize := args[4].I32()
		ident := args[5].I32()

		ret := request_get_field(fieldType, keyPointer, keySize, destPointer, destMaxSize, ident)

		return ret, nil
	}

	return newHostFn("request_get_field", 6, true, fn)
}

func request_get_field(fieldType int32, keyPointer int32, keySize int32, destPointer int32, destMaxSize int32, identifier int32) int32 {
	inst, err := instanceForIdentifier(identifier)
	if err != nil {
		logger.Error(errors.Wrap(err, "[hive-wasm] alert: invalid identifier used, potential malicious activity"))
		return -1
	}

	if inst.request == nil {
		logger.ErrorString("[hive-wasm] Runnable attempted to access request when none is set")
		return -2
	}

	req := inst.request

	keyBytes := inst.readMemory(keyPointer, keySize)
	key := string(keyBytes)

	val := ""

	switch fieldType {
	case fieldTypeMeta:
		switch key {
		case "method":
			val = req.Method
		case "url":
			val = req.URL
		case "id":
			val = req.ID
		case "body":
			val = string(req.Body)
		default:
			return -3
		}
	case fieldTypeBody:
		bodyVal, err := req.BodyField(key)
		if err == nil {
			val = bodyVal
		} else {
			logger.Error(errors.Wrap(err, "failed to get BodyField"))
			return -4
		}
	case fieldTypeHeader:
		header, ok := req.Headers[key]
		if ok {
			val = header
		} else {
			return -3
		}
	case fieldTypeParams:
		param, ok := req.Params[key]
		if ok {
			val = param
		} else {
			return -3
		}
	case fieldTypeState:
		stateVal, ok := req.State[key]
		if ok {
			val = string(stateVal)
		} else {
			return -3
		}
	}

	valBytes := []byte(val)

	if len(valBytes) <= int(destMaxSize) {
		inst.writeMemoryAtLocation(destPointer, valBytes)
	}

	// logger.Debug(fmt.Sprintf("returning value length %d", len(valBytes)))
	return int32(len(valBytes))
}
