package progress

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	. "testing"
	"time"

	. "gopkg.in/check.v1"
)

type progressSuite struct{}

var _ = Suite(&progressSuite{})

func TestProgress(t *T) {
	TestingT(t)
}

func (p *progressSuite) TestMeterReporting(c *C) {
	zero, err := os.Open("/dev/zero")
	c.Assert(err, IsNil)
	r := NewReader("test", zero, 100*time.Millisecond)

	go io.Copy(ioutil.Discard, r)

	now := time.Now()

	var i int
	for tick := range r.C {
		if i > 0 {
			offset := tick.Time.Sub(now)
			fmt.Println(offset)
			c.Assert(offset > time.Duration(i)*100*time.Millisecond && offset < time.Duration(i+1)*100*time.Millisecond, Equals, true)
		}

		c.Assert(tick.Artifact, Equals, "test")

		i++
		if i > 10 {
			r.Close()
		}
	}

	c.Assert(time.Since(now) < 1500*time.Millisecond, Equals, true)
}
