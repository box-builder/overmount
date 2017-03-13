from "golang"

remotepath = "/go/src/github.com/erikh/overmount"

copy ".", remotepath

inside remotepath do
  run "go get -t -v ./..."
end

set_exec entrypoint: [], cmd: ["/bin/sh", "-c", "cd #{remotepath} && go test -v ./... -check.v"]
