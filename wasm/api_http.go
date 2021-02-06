package wasm

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/wasmerio/wasmer-go/wasmer"
)

const (
	methodGet    = int32(1)
	methodPost   = int32(2)
	methodPatch  = int32(3)
	methodDelete = int32(4)
)

const (
	contentTypeJSON        = "application/json"
	contentTypeTextPlain   = "text/plain"
	contentTypeOctetStream = "application/octet-stream"
)

var methodValToMethod = map[int32]string{
	methodGet:    http.MethodGet,
	methodPost:   http.MethodPost,
	methodPatch:  http.MethodPatch,
	methodDelete: http.MethodDelete,
}

func fetchURL() *HostFn {
	fn := func(args ...wasmer.Value) (interface{}, error) {
		method := args[0].I32()
		urlPointer := args[1].I32()
		urlSize := args[2].I32()
		bodyPointer := args[3].I32()
		bodySize := args[4].I32()
		destPointer := args[5].I32()
		destMaxSize := args[6].I32()
		ident := args[7].I32()

		ret := fetch_url(method, urlPointer, urlSize, bodyPointer, bodySize, destPointer, destMaxSize, ident)

		return ret, nil
	}

	return newHostFn("fetch_url", 8, true, fn)
}

func fetch_url(method int32, urlPointer int32, urlSize int32, bodyPointer int32, bodySize int32, destPointer int32, destMaxSize int32, identifier int32) int32 {
	// fetch makes a network request on bahalf of the wasm runner.
	// fetch writes the http response body into memory starting at returnBodyPointer, and the return value is a pointer to that memory
	inst, err := instanceForIdentifier(identifier)
	if err != nil {
		logger.Error(errors.Wrap(err, "[hive-wasm] alert: invalid identifier used, potential malicious activity"))
		return -1
	}

	httpMethod, exists := methodValToMethod[method]
	if !exists {
		logger.ErrorString("invalid method provided")
		return -2
	}

	urlBytes := inst.readMemory(urlPointer, urlSize)

	// the URL is encoded with headers added on the end, each seperated by ::
	// eg. https://google.com/somepage::authorization:bearer qdouwrnvgoquwnrg::anotherheader:nicetomeetyou
	urlParts := strings.Split(string(urlBytes), "::")
	urlString := urlParts[0]

	headers, err := parseHTTPHeaders(urlParts)
	if err != nil {
		logger.Error(errors.Wrap(err, "could not parse URL headers"))
		return -2
	}

	urlObj, err := url.Parse(urlString)
	if err != nil {
		logger.ErrorString("couldn't parse URL")
		return -2
	}

	body := inst.readMemory(bodyPointer, bodySize)

	if len(body) > 0 {
		if headers.Get("Content-Type") == "" {
			headers.Add("Content-Type", contentTypeOctetStream)
		}
	}

	req, err := http.NewRequest(httpMethod, urlObj.String(), bytes.NewBuffer(body))
	if err != nil {
		logger.ErrorString("failed to build request")
		return -2
	}

	req.Header = *headers

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error(errors.Wrap(err, "failed to Do request"))
		return -3
	}

	defer resp.Body.Close()
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.ErrorString("failed to Read response body")
		return -4
	}

	// if the size is greater than what's been allocated, then the module will increase the size and try again
	if len(respBytes) <= int(destMaxSize) {
		inst.writeMemoryAtLocation(destPointer, respBytes)
	}

	return int32(len(respBytes))
}
