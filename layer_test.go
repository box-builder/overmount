package overmount

import (
	"path"

	. "gopkg.in/check.v1"
)

func (m *mountSuite) TestLayerProperties(c *C) {
	layer, err := m.Repository.NewLayer("test", nil)
	c.Assert(err, IsNil)
	_, err = m.Repository.NewLayer("test", nil)
	c.Assert(err, Equals, ErrLayerExists)
	c.Assert(path.Base(layer.MountPath()), Equals, "test")
	c.Assert(path.Dir(layer.MountPath()), Equals, path.Join(m.Repository.baseDir, mountBase))
	c.Assert(path.Base(layer.Path()), Equals, "test")
	c.Assert(path.Dir(layer.Path()), Equals, path.Join(m.Repository.baseDir, layerBase))
}
