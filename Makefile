rs-img = suborbital/hive-rs:$(shell cat ./rs-builder/.rs-ver)

hivew:
	go install ./hivew

test:
	go test -v --count=1 -p=1 ./...

.PHONY: hivew rs-build rs