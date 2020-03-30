# Hive ❤️ WASM

WASM toolchain for Hive

Hive aims for WASM to be a first-class citizen. `hivew` is the central toolchain for Hive and WASM. The `hivew` CLI can build WASM Runnables, and will soon be able to package many WASM Runnables into deployable bundles. It will also be able to act as an all-in-one Hive server, using the upcoming FaaS functionality in Hive.

Docker must be installed to build WASM Runnables.

## Installing
To install `hivew`, clone this repo and run `make hivew`. Go 1.14 must be installed. Package manager installations will be available soon.

## Building Runnables
To build a Rust-based Runnable, see [helloworld-rs](./helloworld-rs/README.md)

`hivew` is under active development alongside [Hive](https://github.com/suborbital/hive) itself.
