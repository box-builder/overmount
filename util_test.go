package overmount

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"

	. "gopkg.in/check.v1"
)

func (m *mountSuite) makeImage(c *C, layerCount int) (*Image, *Layer) {
	var parent *Layer
	for i := 0; i < layerCount; i++ {
		var err error
		parent, err = m.Repository.CreateLayer(fmt.Sprintf("test%d", i), parent)
		c.Assert(err, IsNil)

		r, w := io.Pipe()
		go parent.Unpack(r)

		tw := tar.NewWriter(w)

		wd, err := os.Getwd()
		c.Assert(err, IsNil)

		filepath.Walk(wd, func(p string, fi os.FileInfo, err error) error {
			header, err := tar.FileInfoHeader(fi, fi.Name())
			c.Assert(err, IsNil)
			c.Assert(tw.WriteHeader(header), IsNil)

			if !fi.IsDir() {
				f, err := os.Open(p)
				c.Assert(err, IsNil)
				_, err = io.Copy(tw, f)
				c.Assert(err, IsNil)
				return f.Close()
			}

			return nil
		})

		tw.Close()
	}
	return m.Repository.NewImage(parent), parent
}
