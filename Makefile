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
	docker run -it -v /tmp --privileged --rm erikh/overmount

docker-test:
	go list ./... | grep -v vendor | xargs go test -v -check.v || :
	bash

.PHONY: all test
