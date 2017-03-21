package overmount

import (
	"fmt"

	. "gopkg.in/check.v1"
)

func (m *mountSuite) makeImage(c *C, layerCount int) (*Image, *Layer) {
	var parent *Layer
	for i := 0; i < layerCount; i++ {
		var err error
		parent, err = m.Repository.CreateLayer(fmt.Sprintf("test%d", i), parent)
		c.Assert(err, IsNil)
	}
	return m.Repository.NewImage(parent), parent
}
