package imgio

import (
	"context"
	"io"
	"io/ioutil"
	. "testing"

	. "gopkg.in/check.v1"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	om "github.com/box-builder/overmount"
)

type dockerSuite struct {
	repository *om.Repository
	client     *client.Client
}

var _ = Suite(&dockerSuite{})

func TestImageIO(t *T) {
	TestingT(t)
}

func (d *dockerSuite) SetUpTest(c *C) {
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		panic(err)
	}

	d.repository, err = om.NewRepository(tmpdir)
	if err != nil {
		panic(err)
	}

	d.client, err = client.NewEnvClient()
	if err != nil {
		panic(err)
	}
}

func (d *dockerSuite) TestDockerImport(c *C) {
	images := map[string][]string{
		"golang":          []string{"/bin/bash"}, // should have two images
		"alpine:latest":   nil,                   // squashed image, single layer
		"postgres:latest": []string{"postgres"},  // just a fatty
	}

	for imageName, cmd := range images {
		reader, err := d.client.ImagePull(context.Background(), "docker.io/library/"+imageName, types.ImagePullOptions{})
		c.Assert(err, IsNil, Commentf("%v", imageName))
		_, err = io.Copy(ioutil.Discard, reader)
		c.Assert(err, IsNil, Commentf("%v", imageName))

		reader, err = d.client.ImageSave(context.Background(), []string{imageName})
		c.Assert(err, IsNil, Commentf("%v", imageName))
		docker, err := NewDocker(d.client)
		c.Assert(err, IsNil, Commentf("%v", imageName))
		layers, err := docker.Import(d.repository, reader)
		c.Assert(err, IsNil, Commentf("%v", imageName))
		c.Assert(layers, NotNil, Commentf("%v", imageName))

		for _, layer := range layers {
			config, err := layer.Config()
			c.Assert(err, IsNil, Commentf("%v", imageName))
			c.Assert(config.Config.Cmd, DeepEquals, cmd, Commentf("%v", imageName))

			var count int
			for iter := layer; iter != nil; iter = iter.Parent {
				count++
			}

			c.Assert(count, Equals, len(config.RootFS.DiffIDs))
		}
	}
}
