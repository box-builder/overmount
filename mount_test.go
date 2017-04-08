package overmount

import (
	"github.com/pkg/errors"

	. "gopkg.in/check.v1"
)

func (m *mountSuite) TestMountUnmountProperties(c *C) {
	mount, err := m.Repository.NewMount("/", "", "")
	c.Assert(err, IsNil)
	c.Assert(errors.Cause(mount.Open()), Equals, ErrMountCannotProceed)
}
