# RS-WASM Hive Runnable

To create a Rust-based WASM Runnable, edit `run.rs` in this directory to implement the runnable however you'd like, then from the repo root, call `make rs`, which will run a Docker container that uses `run.rs` and produces `wasm_runnable_bg.wasm`, which can be used with [Hive's WASM support](https://github.com/suborbital/hive/blob/master/WASM.md)!