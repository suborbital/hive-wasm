#! /bin/bash

set -e

cp ../rs-wasm/* ./src/
wasm-pack build

cp pkg/wasm_runner_bg.wasm ../rs-wasm