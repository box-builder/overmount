from "golang"

remotepath = "/go/src/github.com/box-builder/overmount"

copy ".", remotepath

inside remotepath do
  run <<-EOF
  if [ ! -d vendor ]
  then
    set -e
    go get github.com/LK4D4/vndr
    vndr 
  fi
  EOF
end

workdir remotepath

set_exec entrypoint: [], 
         cmd: %w[make docker-test]
