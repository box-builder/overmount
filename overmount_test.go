package overmount

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	. "testing"

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

	p, err := filepath.Rel(m.Repository.BaseDir, t)
	c.Assert(err, IsNil)
	first, _ := path.Split(p)
	c.Assert(err, IsNil)
	c.Assert(first, Equals, tmpdirBase+"/")
}

func (m *mountSuite) TestBasicImageMount(c *C) {
	layerNames := []string{"one", "two", "three"}

	for i := 0; i < len(layerNames); i++ { // offset from 1
		layers := []*Layer{}
		for x, name := range layerNames[:i+1] {
			// stack the layers as parents of each other, except for the first of
			// course.
			var layer *Layer
			if x > 0 {
				layer = layers[x-1]
			}

			layers = append(layers, m.Repository.NewLayer(name, layer))
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

		f, err := os.Create(path.Join(target, image.layer.ID))
		c.Assert(err, IsNil)
		c.Assert(f.Close(), IsNil)

		fis, err := ioutil.ReadDir(target)
		c.Assert(err, IsNil)

		c.Assert(len(fis), Equals, len(layers)) // one file for each layer, one written to each layer

		if len(layers) > 1 {
			c.Assert(image.Unmount(), IsNil)
		}
	}
}
