#! /bin/bash

cp ../rs-wasm/run.rs ./src/
wasm-pack build

cp pkg/wasm_runner_bg.wasm ../rs-wasm