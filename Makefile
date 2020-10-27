rs-img = suborbital/hive-rs:$(shell cat ./rs-builder/.rs-ver)

test:
	go test -v --count=1 -p=1 ./...

rs-raw-wasm:
	cp ../subo/builders/rust/target/wasm32-wasi/release/hivew_rs_builder.wasm ./wasm/testdata/

swift-raw-wasm:
	cp ../subo/builders/swift/runnable.wasm ./wasm/testdata/swiftc_runnable.wasm

.PHONY: hivew rs-build rs