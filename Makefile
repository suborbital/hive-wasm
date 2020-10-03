rs-img = suborbital/hive-rs:$(shell cat ./rs-builder/.rs-ver)

hivew:
	go install ./hivew

test:
	go test -v --count=1 -p=1 ./...

rs-raw-wasm:
	cp ../hivew-rs-builder/target/wasm32-wasi/release/hivew_rs_builder.wasm ./wasm/testdata/

.PHONY: hivew rs-build rs