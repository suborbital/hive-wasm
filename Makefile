rs-img = suborbital/hive-rs:$(shell cat ./rs-builder/.rs-ver)

rs-build:
	docker build . -f rs-builder/Dockerfile -t $(rs-img)

rs:
	docker run -it --mount type=bind,source="$(PWD)"/rs-wasm,target=/root/rs-wasm $(rs-img)