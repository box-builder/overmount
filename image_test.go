package overmount

import (
	"github.com/pkg/errors"
	. "gopkg.in/check.v1"
)

func (m *mountSuite) TestImageMountUnmount(c *C) {
	image, layer := m.makeImage(c, 2)

	image2 := m.Repository.NewImage(layer.parent) // only one layer
	err := image2.Mount()
	c.Assert(errors.Cause(err), Equals, ErrMountCannotProceed)
	c.Assert(image2.Unmount(), Equals, ErrMountCannotProceed)
	c.Assert(image.Mount(), IsNil)
	c.Assert(image.Unmount(), IsNil)
}
