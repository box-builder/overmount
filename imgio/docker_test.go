package imgio

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"strings"
	. "testing"

	. "gopkg.in/check.v1"

	om "github.com/box-builder/overmount"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
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

	d.repository, err = om.NewRepository(tmpdir, os.Getenv("VIRTUAL") != "")
	if err != nil {
		panic(err)
	}

	d.client, err = client.NewEnvClient()
	if err != nil {
		panic(err)
	}
}

func (d *dockerSuite) TestDockerImportExport(c *C) {
	images := map[string][]string{
		"golang":          []string{"/bin/bash"}, // should have two images
		"alpine:latest":   nil,                   // squashed image, single layer
		"postgres:latest": []string{"postgres"},  // just a fatty
	}

	layerIDMap := map[string]struct{}{}

	tags := map[string]struct{}{
		"golang:latest":   struct{}{},
		"alpine:latest":   struct{}{},
		"postgres:latest": struct{}{},
	}

	docker, err := NewDocker(d.client)
	c.Assert(err, IsNil)

	for imageName, cmd := range images {
		reader, err := d.client.ImagePull(context.Background(), "docker.io/library/"+imageName, types.ImagePullOptions{})
		c.Assert(err, IsNil, Commentf("%v", imageName))
		_, err = io.Copy(ioutil.Discard, reader)
		c.Assert(err, IsNil, Commentf("%v", imageName))

		reader, err = d.client.ImageSave(context.Background(), []string{imageName})
		c.Assert(err, IsNil, Commentf("%v", imageName))
		layers, err := docker.Import(d.repository, reader)
		c.Assert(err, IsNil, Commentf("%v", imageName))
		c.Assert(layers, NotNil, Commentf("%v", imageName))

		for _, layer := range layers {
			layerIDMap[layer.ID()] = struct{}{}
			config, err := layer.Config()
			c.Assert(err, IsNil, Commentf("%v", imageName))
			c.Assert(config.Cmd, DeepEquals, cmd, Commentf("%v", imageName))

			reader, err = d.repository.Export(docker, layer)
			c.Assert(err, IsNil, Commentf("%v", imageName))
			resp, err := d.client.ImageLoad(context.Background(), reader, false)
			c.Assert(err, IsNil, Commentf("%v", imageName))

			br := bufio.NewReader(resp.Body)
			for {
				line, err := br.ReadString('\n')
				line = strings.TrimSpace(line)
				if line != "" {
					myMap := map[string]interface{}{}
					c.Assert(json.Unmarshal([]byte(line), &myMap), IsNil)
					c.Assert(myMap["stream"], NotNil)
					c.Assert(strings.HasPrefix(myMap["stream"].(string), "Loaded image ID"), Equals, true)
				}
				if err != nil {
					break
				}
			}

			c.Assert(err, IsNil, Commentf("%v", imageName))
		}
	}

	for tag := range tags {
		layer, err := d.repository.GetTag(tag)
		c.Assert(err, IsNil)
		_, ok := layerIDMap[layer.ID()]
		c.Assert(ok, Equals, true)
	}
}
