# Hive ❤️ WASM

`hive-wasm` is the WASM API for [Hive](https://github.com/suborbital/hive).

These packages are mostly used by Hive internally to run Wasm modules, and include the Hive Wasm FFI API, as well as the Runnable implementation for Wasm modules. 

`hive-wasm` includes a multi-tenant Wasm runtime (powered by [Wasmer](https://wasmer.io) under the hood) which allows many modules to be running simultaneously with inputs, outputs, function calls, and memory management handled automatically.

Building Wasm Runnables is done using the [subo CLI](https://github.com/suborbital/subo).