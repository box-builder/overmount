package overmount

import (
	"io/ioutil"

	. "gopkg.in/check.v1"
)

func (m *mountSuite) TestBasicRepository(c *C) {
	tempdir, err := ioutil.TempDir("", "")
	c.Assert(err, IsNil)
	r, err := NewRepository(tempdir, false)
	c.Assert(err, IsNil)
	c.Assert(r.baseDir, Equals, tempdir)
	c.Assert(r.editMutex, NotNil)
	c.Assert(r.layers, NotNil)
	c.Assert(r.mounts, NotNil)
	_, err = NewRepository("/dev/zero", false)
	c.Assert(err, NotNil)

	r.baseDir = "/dev/zero"
	_, err = r.TempDir()
	c.Assert(err, NotNil)
}

func (m *mountSuite) TestRepositoryPropagation(c *C) {
	if m.Repository.IsVirtual() {
		c.Skip("Cannot mount virtual layers")
		return
	}

	image, layer := m.makeImage(c, 2)
	c.Assert(image.Mount(), IsNil)
	c.Assert(len(m.Repository.mounts), Equals, 1)
	c.Assert(len(m.Repository.layers), Equals, 2)
	c.Assert(image.Unmount(), IsNil)

	m.Repository.RemoveLayer(layer)
	c.Assert(len(m.Repository.layers), Equals, 1)
	m.Repository.RemoveLayer(layer.Parent)
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

	c.Assert(m.Repository.AddLayer(layer, false), IsNil)
	c.Assert(m.Repository.AddLayer(layer, false), Equals, ErrLayerExists)
	c.Assert(m.Repository.AddLayer(layer, true), IsNil)
	c.Assert(len(m.Repository.layers), Equals, 1)
	c.Assert(m.Repository.AddLayer(layer.Parent, false), IsNil)
	c.Assert(m.Repository.AddLayer(layer.Parent, false), Equals, ErrLayerExists)
	c.Assert(m.Repository.AddLayer(layer.Parent, true), IsNil)
	c.Assert(len(m.Repository.layers), Equals, 2)
	m.Repository.RemoveLayer(layer)
	c.Assert(len(m.Repository.layers), Equals, 1)
	m.Repository.RemoveLayer(layer.Parent)
	c.Assert(len(m.Repository.layers), Equals, 0)
}
