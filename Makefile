all: test

clean:
	rm -rf vendor

vndr:
	go get github.com/LK4D4/vndr
	vndr

install_box:
	@sh install_box.sh

build: install_box
	box -t box-builder/overmount build.rb

test: build
	make run-docker

test-virtual: build
	VIRTUAL=1 make run-docker

test-ci:
	@sh install_box_ci.sh
	bin/box -t box-builder/overmount build.rb
	make run-docker
	VIRTUAL=1 make run-docker

run-docker:
	docker run -e VIRTUAL=${VIRTUAL} -v /var/run/docker.sock:/var/run/docker.sock -v /tmp --privileged --rm box-builder/overmount

docker-deps:
	go get github.com/opencontainers/image-tools/...

docker-test: docker-deps
	go build -v -o /dev/null ./examples/... ./om/...
	set -e; for test in $(shell go list ./... | grep -v vendor); do go test -cover -v $${test} -check.v; done

.PHONY: test
