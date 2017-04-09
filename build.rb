from "golang"

remotepath = "/go/src/github.com/box-builder/overmount"

copy ".", remotepath

inside remotepath do
  run "chmod 777 ."

  run <<-EOF
  if [ ! -d vendor ]
  then
    set -e
    go get github.com/LK4D4/vndr
    vndr 
  fi
  EOF
end

run "groupadd -g 999 docker"
run "usermod -aG docker nobody"
run "chmod -R 777 /go"

workdir remotepath

set_exec entrypoint: [], 
         cmd: %w[make docker-test]
