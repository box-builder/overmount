package imgio

import (
	"context"
	"io"
	"io/ioutil"
	. "testing"

	. "gopkg.in/check.v1"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	om "github.com/erikh/overmount"
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
	reader, err := d.client.ImagePull(context.Background(), "docker.com/library/golang:latest", types.ImagePullOptions{})
	c.Assert(err, IsNil)
	_, err = io.Copy(ioutil.Discard, reader)
	c.Assert(err, IsNil)

	reader, err = d.client.ImageSave(context.Background(), []string{"golang:latest"})
	c.Assert(err, IsNil)
	docker, err := NewDocker(nil)
	c.Assert(err, IsNil)
	layer, err := docker.Import(d.repository, reader)
	c.Assert(err, IsNil)
	c.Assert(layer, NotNil)
	config, err := layer.Config()
	c.Assert(err, IsNil)
	c.Assert(config.WorkingDir, Equals, "/go")

	inspect, _, err := d.client.ImageInspectWithRaw(context.Background(), "golang:latest")
	c.Assert(err, IsNil)

	var count int
	for iter := layer; iter != nil; iter = iter.Parent {
		count++
	}

	c.Assert(count, Equals, len(inspect.RootFS.Layers))
}
