package imgio

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	. "gopkg.in/check.v1"

	"github.com/docker/docker/api/types"
)

func (d *dockerSuite) TestOCIExport(c *C) {
	images := map[string][]string{
		"golang":          []string{"/bin/bash"}, // should have two images
		"alpine:latest":   nil,                   // squashed image, single layer
		"postgres:latest": []string{"postgres"},  // just a fatty
	}

	docker, err := NewDocker(d.client)
	c.Assert(err, IsNil)

	for imageName := range images {
		reader, err := d.client.ImagePull(context.Background(), "docker.io/library/"+imageName, types.ImagePullOptions{})
		c.Assert(err, IsNil, Commentf("%v", imageName))
		_, err = io.Copy(ioutil.Discard, reader)
		c.Assert(err, IsNil, Commentf("%v", imageName))

		reader, err = d.client.ImageSave(context.Background(), []string{imageName})
		c.Assert(err, IsNil, Commentf("%v", imageName))
		layers, err := docker.Import(d.repository, reader)
		c.Assert(err, IsNil, Commentf("%v", imageName))
		c.Assert(layers, NotNil, Commentf("%v", imageName))

		oci := NewOCI()
		reader, err = d.repository.Export(oci, layers[0], []string{"oci"})
		c.Assert(err, IsNil)
		tf, err := ioutil.TempFile("", "overmount-test-")
		defer os.Remove(tf.Name())
		c.Assert(err, IsNil)
		_, err = io.Copy(tf, reader)
		c.Assert(err, IsNil)
		tf.Close()

		out, err := exec.Command("oci-image-tool", "validate", "--ref", "oci", tf.Name()).CombinedOutput()
		c.Assert(err, IsNil)
		c.Assert(strings.Contains(string(out), fmt.Sprintf("%s: OK", tf.Name())), Equals, true)
		c.Assert(strings.Contains(string(out), "Validation succeeded"), Equals, true)
	}
}
