all: test

clean:
	rm -rf vendor

vndr:
	go get github.com/LK4D4/vndr
	vndr

install_box:
	@sh install_box.sh

test: install_box
	box -t erikh/overmount build.rb
	make run-docker

test-ci:
	@sh install_box_ci.sh
	bin/box -t erikh/overmount build.rb
	make run-docker

run-docker:
	docker run -v /var/run/docker.sock:/var/run/docker.sock -v /tmp --privileged --rm erikh/overmount

docker-test:
	go build -v -o /dev/null ./examples/... ./om/...
	set -e; for test in $(shell go list ./... | grep -v vendor); do go test -cover -v $${test} -check.v; done

.PHONY: test
