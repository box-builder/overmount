all: test

install_box:
	@sh install_box.sh

test: install_box
	box -t erikh/overmount build.rb	
	docker run -it --rm erikh/overmount

.PHONY: all test
