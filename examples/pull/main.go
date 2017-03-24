//
// this is a VERY quick & dirty docker image unpacker. It will unpack each
// layer into its own path and then join them with the overlayfs mount at the
// end.
//

package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/erikh/overmount"
	"github.com/erikh/overmount/imgio"
)

func main() {
	if os.Getuid() != 0 {
		panic("you need to be root to run this thing; sorry!")
	}

	var tmpdir string

	if len(os.Args) == 2 {
		tmpdir = os.Args[1]
	} else {
		var err error

		tmpdir, err = ioutil.TempDir("", "")
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("repository:", tmpdir)

	repo, err := overmount.NewRepository(tmpdir)
	if err != nil {
		panic(err)
	}

	dockerClient, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	reader, err := dockerClient.ImagePull(context.Background(), "docker.io/library/golang:latest", types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(ioutil.Discard, reader)
	if err != nil {
		panic(err)
	}

	reader, err = dockerClient.ImageSave(context.Background(), []string{"library/golang:latest"})
	if err != nil {
		panic(err)
	}

	docker, err := imgio.NewDocker(dockerClient)
	if err != nil {
		panic(err)
	}

	layers, err := repo.Import(docker, reader)
	if err != nil {
		panic(err)
	}

	for _, layer := range layers {
		fmt.Println(layer.ID())
	}
}
