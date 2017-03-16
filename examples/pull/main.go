//
// this is a VERY quick & dirty docker image unpacker. It will unpack each
// layer into its own path and then join them with the overlayfs mount at the
// end.
//

package main

import (
	"archive/tar"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/erikh/overmount"
)

func main() {
	if os.Getuid() != 0 {
		panic("you need to be root to run this thing; sorry!")
	}

	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		panic(err)
	}
	repo, err := overmount.NewRepository(tmpdir)
	if err != nil {
		panic(err)
	}

	dockerClient, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	reader, err := dockerClient.ImagePull(context.Background(), "docker.com/library/golang:latest", types.ImagePullOptions{})
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

	layerMap := map[string]string{}
	var manifest []map[string]interface{}

	tr := tar.NewReader(reader)

	tmpdir2, err := ioutil.TempDir("", "")
	if err != nil {
		panic(err)
	}

	for {
		header, err := tr.Next()
		if err != nil {
			break
		}
		if path.Base(header.Name) == "layer.tar" {
			layerID := path.Base(path.Dir(header.Name))
			f, err := os.Create(path.Join(tmpdir2, layerID+".tar"))
			if err != nil {
				panic(err)
			}
			if _, err := io.Copy(f, tr); err != nil {
				panic(err)
			}
			f.Close()
			layerMap[layerID] = f.Name()
		} else if path.Base(header.Name) == "manifest.json" {
			content, err := ioutil.ReadAll(tr)
			if err != nil {
				panic(err)
			}
			if err := json.Unmarshal(content, &manifest); err != nil {
				panic(err)
			}
		} else {
			io.Copy(ioutil.Discard, tr)
		}
	}
	reader.Close()

	var parent *overmount.Layer // at the end of the loop, the parent will be the top-most layer

	for _, tmp := range manifest[0]["Layers"].([]interface{}) {
		layerID := path.Dir(tmp.(string))
		tarfile, err := os.Open(layerMap[layerID])
		if err != nil {
			panic(err)
		}
		layer, err := repo.NewLayer(layerID, parent)
		if err != nil {
			panic(err)
		}
		if err := os.MkdirAll(layer.Path(), 0700); err != nil {
			panic(err)
		}
		parent = layer
		digest, err := layer.Unpack(tarfile)
		if err != nil {
			panic(err)
		}

		fmt.Printf("unpacked layer with digest %q (id: %v) to %v\n", digest[7:19], layer.ID[:12], layer.Path())
	}

	image := repo.NewImage(parent)
	if err := image.Mount(); err != nil {
		panic(err)
	}

	fmt.Printf("Mount is at %v\n", parent.MountPath())
}
