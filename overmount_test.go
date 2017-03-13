package overmount

import (
	. "testing"

	. "gopkg.in/check.v1"
)

type mountSuite struct{}

var _ = Suite(&mountSuite{})

func TestOvermount(t *T) {
	TestingT(t)
}
