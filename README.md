# hive-wasm

WASM bundle manager for Hive

This project contains the tools and libraries needed to build WASM Runnables for [Hive](https://github.com/suborbital/hive). Each supported language has a Dockerfile and associated boilerplate that are used to build your runnables. This project is very early days and can currently manually build Rust-based Runnables.

Docker must be installed to build WASM Runnables. You must also clone this repo and have `make` available.

To build a Rust-based WASM runnable, see [rs-wasm](./rs-wasm/README.md)

The plan for this repo is to build a tool called `hivew` which uses official Docker images to build your runnables without needing to have this repo cloned.