all: test

clean:
	rm -rf vendor

vndr:
	go get github.com/LK4D4/vndr
	vndr

install_box:
	@sh install_box.sh

test: install_box
	$(shell which box) -t erikh/overmount build.rb	
	docker run -v /var/run/docker.sock:/var/run/docker.sock -it -v /tmp --privileged --rm erikh/overmount

docker-test:
	go build -v -o /dev/null ./examples/... 
	go list ./... | grep -v vendor | xargs go test -cover -v -check.v

.PHONY: test
