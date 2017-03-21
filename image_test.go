package overmount

import (
	"os"

	. "gopkg.in/check.v1"

	"github.com/pkg/errors"
)

func (m *mountSuite) TestImageMountUnmount(c *C) {
	image, layer := m.makeImage(c, 2)

	image2 := m.Repository.NewImage(layer.Parent()) // only one layer
	err := image2.Mount()
	c.Assert(errors.Cause(err), Equals, ErrMountCannotProceed)
	c.Assert(image2.Unmount(), Equals, ErrMountCannotProceed)
	c.Assert(image.Mount(), IsNil)
	c.Assert(image.Unmount(), IsNil)
}

func (m *mountSuite) TestImageCommit(c *C) {
	image, layer := m.makeImage(c, 10)
	for iter := layer; iter != nil; iter = iter.Parent() {
		_, err := os.Stat(iter.parentsPath())
		c.Assert(err, NotNil)
	}

	c.Assert(image.Commit(), IsNil)

	for iter := layer; iter != nil; iter = iter.Parent() {
		if iter.Parent() != nil {
			_, err := os.Stat(iter.parentsPath())
			c.Assert(err, IsNil)
		}
	}
}
