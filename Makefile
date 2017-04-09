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

test-cover: build run-docker-cover
	
test-virtual-cover: build
	VIRTUAL=1 make run-docker-cover

test: build run-docker

test-virtual: build
	VIRTUAL=1 make run-docker

test-unprivileged: build
	UNPRIVILEGED=1 VIRTUAL=1 make run-docker

test-unprivileged-cover: build
	UNPRIVILEGED=1 VIRTUAL=1 make run-docker-cover

test-ci:
	@sh install_box_ci.sh
	bin/box -t box-builder/overmount build.rb
	make run-docker
	VIRTUAL=1 make run-docker
	UNPRIVILEGED=1 VIRTUAL=1 make run-docker

TEST_NAME:="overmount-test-$(shell head -c +10 /dev/urandom | sha256sum | awk '{ print $$1 }')"
MYUID := $(if $(strip ${UNPRIVILEGED}),nobody,root)

run-docker-cover:
	docker run -u ${MYUID}:docker -e VIRTUAL=${VIRTUAL} -e UNPRIVILEGED=${UNPRIVILEGED} -v /var/run/docker.sock:/var/run/docker.sock -v /tmp --privileged --name ${TEST_NAME} box-builder/overmount
	@echo Your test container ID is ${TEST_NAME}

run-docker:
	docker run -u ${MYUID}:docker -e VIRTUAL=${VIRTUAL} -e UNPRIVILEGED=${UNPRIVILEGED} -v /var/run/docker.sock:/var/run/docker.sock -v /tmp --privileged --rm box-builder/overmount

docker-deps:
	go get github.com/opencontainers/image-tools/...

docker-test: docker-deps
	go build -v -o /dev/null ./examples/... ./om/...
	set -e; for test in $(shell go list ./... | grep -v vendor); do go test -coverprofile $$(echo $${test} | sed -e 's!^github.com/box-builder/overmount!$(shell pwd)!')/cover.out -v $${test} -check.v; done

.PHONY: test
