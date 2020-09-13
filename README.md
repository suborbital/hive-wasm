# Hive ❤️ WASM

WASM toolchain for [Hive](https://github.com/suborbital/hive)

The Hive job scheduler aims for WASM to be a first-class citizen. `hivew` is the toolchain for Hive and WASM. The `hivew` CLI can build WASM Runnables, and can package many WASM Runnables into a deployable bundle. It will soon be able to act as an all-in-one Hive server, using Hive's FaaS functionality.

Docker must be installed to build WASM Runnables.

## Installing
To install `hivew`, clone this repo and run `make hivew`. A version of Go that supports Modules is required. Package manager installations will be available soon.

You can also install using [gobinaries](https://gobinaries.com/):
```
curl -sf https://gobinaries.com/suborbital/hivew/hivew | sh
```

## Building Runnables
To build a Rust-based Runnable, see [helloworld-rs](./helloworld-rs/README.md)

## Bundles
To build all of the Runnables in the current directory and bundle them all into a single `.wasm.zip` file, run `hivew build --bundle`. The resulting bundle can be used with a Hive instance by calling `h.HandleBundle({path/to/bundle})`. See the [hive WASM instructions](https://github.com/suborbital/hive/blob/master/WASM.md) for details.

`hivew` is under active development alongside [Hive](https://github.com/suborbital/hive) itself.

Copyright Suborbital contributors 2020
