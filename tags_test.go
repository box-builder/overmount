package overmount

import (
	"github.com/pkg/errors"
	. "gopkg.in/check.v1"
)

func (m *mountSuite) TestTags(c *C) {
	_, err := m.Repository.GetTag("test")
	c.Assert(errors.Cause(err), Equals, ErrTagDoesNotExist)

	err = m.Repository.RemoveTag("test")
	c.Assert(errors.Cause(err), Equals, ErrTagDoesNotExist)

	_, layer := m.makeImage(c, 2)
	c.Assert(m.Repository.AddTag("test", layer), IsNil)

	layer2, err := m.Repository.GetTag("test")
	c.Assert(err, IsNil)

	c.Assert(layer2.ID(), Equals, layer.ID())
	c.Assert(layer2.RestoreParent(), IsNil)
	c.Assert(layer2.Parent.ID(), Equals, layer.Parent.ID())
	c.Assert(m.Repository.RemoveTag("test"), IsNil)
}
