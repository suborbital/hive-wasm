# This Dockerfile (and everything in rs-builder) is internal to the tool and doesn't need to be changed.

FROM rust:1.42-stretch

RUN curl https://rustwasm.github.io/wasm-pack/installer/init.sh -sSf | sh

WORKDIR /root
RUN mkdir rs-builder; mkdir rs-wasm
WORKDIR /root/rs-builder

ADD ./rs-builder .

RUN wasm-pack build
RUN rm ./src/run.rs; rm -r ./pkg

ENTRYPOINT ./build.sh
