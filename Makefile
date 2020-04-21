rs-img = suborbital/hive-rs:$(shell cat ./rs-builder/.rs-ver)

hivew:
	go install ./hivew

test:
	go test -v --count=1 -p=1 ./...

rs-build:
	docker build . -f rs-builder/Dockerfile -t $(rs-img)

rs:
	docker run -it --mount type=bind,source="$(PWD)"/rs-wasm,target=/root/rs-wasm $(rs-img)

.PHONY: hivew rs-build rs