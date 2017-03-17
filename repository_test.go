package overmount

import (
	"io/ioutil"

	. "gopkg.in/check.v1"
)

func (m *mountSuite) TestBasicRepository(c *C) {
	tempdir, err := ioutil.TempDir("", "")
	c.Assert(err, IsNil)
	r, err := NewRepository(tempdir)
	c.Assert(err, IsNil)
	c.Assert(r.baseDir, Equals, tempdir)
	c.Assert(r.editMutex, NotNil)
	c.Assert(r.layers, NotNil)
	c.Assert(r.mounts, NotNil)
	_, err = NewRepository("/dev/zero")
	c.Assert(err, NotNil)

	r.baseDir = "/dev/zero"
	_, err = r.TempDir()
	c.Assert(err, NotNil)
}

func (m *mountSuite) TestRepositoryPropagation(c *C) {
	image, layer := m.makeImage(c, 2)
	c.Assert(image.Mount(), IsNil)
	c.Assert(len(m.Repository.mounts), Equals, 1)
	c.Assert(len(m.Repository.layers), Equals, 2)
	c.Assert(image.Unmount(), IsNil)

	m.Repository.RemoveLayer(layer)
	c.Assert(len(m.Repository.layers), Equals, 1)
	m.Repository.RemoveLayer(layer.parent)
	c.Assert(len(m.Repository.layers), Equals, 0)

	c.Assert(len(m.Repository.mounts), Equals, 1)
	// XXX we can't normally access the mount so we cheat here.
	m.Repository.RemoveMount(image.mount)
	c.Assert(len(m.Repository.mounts), Equals, 0)

	// XXX here too
	c.Assert(m.Repository.AddMount(image.mount), IsNil)
	c.Assert(len(m.Repository.mounts), Equals, 1)
	m.Repository.RemoveMount(image.mount)
	c.Assert(len(m.Repository.mounts), Equals, 0)

	c.Assert(m.Repository.AddLayer(layer), IsNil)
	c.Assert(m.Repository.AddLayer(layer), Equals, ErrLayerExists)
	c.Assert(len(m.Repository.layers), Equals, 1)
	c.Assert(m.Repository.AddLayer(layer.parent), IsNil)
	c.Assert(m.Repository.AddLayer(layer.parent), Equals, ErrLayerExists)
	c.Assert(len(m.Repository.layers), Equals, 2)
	m.Repository.RemoveLayer(layer)
	c.Assert(len(m.Repository.layers), Equals, 1)
	m.Repository.RemoveLayer(layer.parent)
	c.Assert(len(m.Repository.layers), Equals, 0)
}
