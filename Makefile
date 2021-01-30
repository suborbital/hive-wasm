test:
	go test -v -count=1 ./...

test/data:
	subo build ./wasm/testdata --native

.PHONY: test