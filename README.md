# Hive ❤️ WASM

WASM toolchain for [Hive](https://github.com/suborbital/hive)

The Hive job scheduler aims for WASM to be a first-class citizen. `hivew` is the central toolchain for Hive and WASM. The `hivew` CLI can build WASM Runnables, and can package many WASM Runnables into deployable bundles. It will soon be able to act as an all-in-one Hive server, using the upcoming FaaS functionality in Hive.

Docker must be installed to build WASM Runnables.

## Installing
To install `hivew`, clone this repo and run `make hivew`. Go 1.14 must be installed. Package manager installations will be available soon.

## Building Runnables
To build a Rust-based Runnable, see [helloworld-rs](./helloworld-rs/README.md)

## Bundles
To build all of the Runnables in the current directory and bundle them all into a single `.wasm.zip` file, run `hivew build --bundle`. The resulting bundle can be used with a Hive instance by calling `h.HandleBundle({path/to/bundle})`. See the [hive WASM instructions](https://github.com/suborbital/hive/blob/master/WASM.md) for details.

`hivew` is under active development alongside [Hive](https://github.com/suborbital/hive) itself.

Copyright Suborbital contributors 2020
