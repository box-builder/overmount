package overmount

import (
	"archive/tar"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	. "testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	. "gopkg.in/check.v1"
)

type mountSuite struct {
	Repository *Repository
}

var _ = Suite(&mountSuite{})

func TestOvermount(t *T) {
	TestingT(t)
}

func (m *mountSuite) SetUpTest(c *C) {
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		panic(err)
	}
	repo, err := NewRepository(tmpdir)
	if err != nil {
		panic(err)
	}

	m.Repository = repo
}

func (m *mountSuite) TestRepositoryTempDir(c *C) {
	t, err := m.Repository.TempDir()
	c.Assert(err, IsNil)

	p, err := filepath.Rel(m.Repository.baseDir, t)
	c.Assert(err, IsNil)
	first, _ := path.Split(p)
	c.Assert(err, IsNil)
	c.Assert(first, Equals, tmpdirBase+"/")
}

func (m *mountSuite) TestBasicImageMount(c *C) {
	layerNames := []string{"one", "two", "three"}

	for i := 0; i < len(layerNames); i++ {
		layers := []*Layer{}
		for x, name := range layerNames[:i+1] {
			// stack the layers as parents of each other, except for the first of
			// course.
			var parent *Layer
			if x > 0 {
				parent = layers[x-1]
			}

			child, err := m.Repository.NewLayer(name, parent)
			c.Assert(err, IsNil)
			layers = append(layers, child)
		}

		image := m.Repository.NewImage(layers[len(layers)-1])
		if len(layers) == 1 {
			c.Assert(image.Mount(), NotNil)
			m.Repository.mkdirCheckRel(image.layer.Path())
		} else {
			c.Assert(image.Mount(), IsNil)
			c.Assert(image.mount.Mounted(), Equals, true)
		}

		target := image.layer.MountPath()
		if len(layers) == 1 {
			target = image.layer.Path()
		}

		r, w, err := os.Pipe()
		c.Assert(err, IsNil)
		errChan := make(chan error, 1)

		go func(target string) {
			tw := tar.NewWriter(w)
			defer w.Close()
			defer tw.Close()
			defer close(errChan)

			err = tw.WriteHeader(&tar.Header{
				Name:     image.layer.ID(),
				Mode:     0600,
				Typeflag: tar.TypeReg,
			})
			if err != nil {
				errChan <- err
				return
			}
			if _, err := tw.Write([]byte{}); err != nil {
				errChan <- err
				return
			}
		}(target)

		c.Assert(image.layer.asset.Read(r), IsNil)
		fis, err := ioutil.ReadDir(target)
		c.Assert(err, IsNil)

		c.Assert(len(fis), Equals, len(layers)) // one file for each layer, one written to each layer

		if len(layers) > 1 {
			c.Assert(image.Unmount(), IsNil)
		}

		for _, layer := range layers {
			m.Repository.RemoveLayer(layer)
		}
	}
}

func (m *mountSuite) TestImageUnpack(c *C) {
	dockerClient, err := client.NewEnvClient()
	c.Assert(err, IsNil)

	reader, err := dockerClient.ImagePull(context.Background(), "docker.com/library/golang:latest", types.ImagePullOptions{})
	c.Assert(err, IsNil)
	_, err = io.Copy(ioutil.Discard, reader)
	c.Assert(err, IsNil)

	reader, err = dockerClient.ImageSave(context.Background(), []string{"library/golang:latest"})
	c.Assert(err, IsNil)

	layerMap := map[string]string{}
	var manifest []map[string]interface{}

	tr := tar.NewReader(reader)

	tmpdir, err := ioutil.TempDir("", "")
	c.Assert(err, IsNil)

	for {
		header, err := tr.Next()
		if err != nil {
			break
		}
		if path.Base(header.Name) == "layer.tar" {
			layerID := path.Base(path.Dir(header.Name))
			f, err := os.Create(path.Join(tmpdir, layerID+".tar"))
			c.Assert(err, IsNil)
			_, err = io.Copy(f, tr)
			c.Assert(err, IsNil)
			f.Close()
			layerMap[layerID] = f.Name()
		} else if path.Base(header.Name) == "manifest.json" {
			content, err := ioutil.ReadAll(tr)
			c.Assert(err, IsNil)
			c.Assert(json.Unmarshal(content, &manifest), IsNil)
		} else {
			io.Copy(ioutil.Discard, tr)
		}
	}
	reader.Close()

	var parent *Layer // at the end of the loop, the parent will be the top-most layer

	for _, tmp := range manifest[0]["Layers"].([]interface{}) {
		layerID := path.Dir(tmp.(string))
		tarfile, err := os.Open(layerMap[layerID])
		c.Assert(err, IsNil)
		layer, err := m.Repository.NewLayer(layerID, parent)
		c.Assert(os.MkdirAll(layer.Path(), 0700), IsNil)
		c.Assert(err, IsNil)
		parent = layer
		digest, err := layer.Unpack(tarfile)
		c.Assert(err, IsNil)
		c.Assert(digest, NotNil)
	}

	image := m.Repository.NewImage(parent)
	c.Assert(image.Mount(), IsNil)
	_, err = os.Stat(path.Join(parent.MountPath(), "/usr/local/go/bin/go"))
	c.Assert(err, IsNil)
	_, err = os.Stat(path.Join(parent.MountPath(), "/go/bin"))
	c.Assert(err, IsNil)
	c.Assert(image.Unmount(), IsNil)
	_, err = os.Stat(path.Join(parent.MountPath(), "/usr/local/go/bin/go"))
	c.Assert(err, NotNil)
	_, err = os.Stat(path.Join(parent.MountPath(), "/go/bin"))
	c.Assert(err, NotNil)
}
