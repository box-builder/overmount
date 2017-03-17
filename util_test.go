package overmount

import (
	"fmt"

	. "gopkg.in/check.v1"
)

func (m *mountSuite) makeImage(c *C, layerCount int) (*Image, *Layer) {
	var parent *Layer
	for i := 0; i < layerCount; i++ {
		layer, err := m.Repository.NewLayer(fmt.Sprintf("test%d", i), parent)
		c.Assert(err, IsNil)
		parent = layer
	}
	return m.Repository.NewImage(parent), parent
}
