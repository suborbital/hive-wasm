# Hive ❤️ WASM

WASM toolchain for [Hive](https://github.com/suborbital/hive)

The Hive job scheduler aims for WASM to be a first-class citizen. `hivew` is the toolchain for Hive and WASM. The `hivew` CLI can build WASM Runnables, and can package many WASM Runnables into a deployable bundle. It will soon be able to act as an all-in-one Hive server, using Hive's FaaS functionality.

Writing a Runnable for Hive in languages other than Go is designed to be just as simple and powerful:
```rust
#[no_mangle]
pub fn run(input: Vec<u8>) -> Option<Vec<u8>> {
    let in_string = String::from_utf8(input).unwrap();

    Some(String::from(format!("hello {}", in_string)).as_bytes().to_vec())
}
```
hivew will package your Runnable into a WASM module that can be loaded into a Hive instance and run just like any other Runnable!


## Installing
To install `hivew`, clone this repo and run `make hivew`. A version of Go that supports Modules is required. Package manager installations will be available soon.

You can also install using [gobinaries](https://gobinaries.com/):
```
curl -sf https://gobinaries.com/suborbital/hivew/hivew | sh
```

## Building WASM Runnables
**Docker must be installed to build WASM Runnables.**
The hivew CLI builds your Runnable code into a WASM module that can be loaded by Hive.

To build a Rust-based Runnable, see [helloworld-rs](./helloworld-rs/README.md)

## Bundles
To build all of the Runnables in the current directory and bundle them all into a single `.wasm.zip` file, run `hivew build --bundle`. The resulting bundle can be used with a Hive instance by calling `h.HandleBundle({path/to/bundle})`. See the [hive WASM instructions](https://github.com/suborbital/hive/blob/master/WASM.md) for details.

`hivew` is under active development alongside [Hive](https://github.com/suborbital/hive) itself.

Copyright Suborbital contributors 2020

## FFI Runnable API
hivew provides an API which allows for communication between WASM runnables and Hive. This API is currently limited to internal functions used to run jobs and return their results, but in the future this API will expand to include:
- The ability to make network requests from WASM Runnables (with built-in access controls to restrict network activity)
- The ability to read files from the host machine (with build-in access control)
- The ability to schedule new Hive jobs and get their results (similar to the Go Runnable API)