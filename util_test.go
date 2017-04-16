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
		parent, err = m.Repository.CreateLayer(fmt.Sprintf("test%d", i), parent, false)
		c.Assert(err, IsNil)

		r, w := io.Pipe()
		go func() {
			_, err := parent.Unpack(r)
			c.Assert(err, IsNil)
		}()

		tw := tar.NewWriter(w)

		wd, err := os.Getwd()
		c.Assert(err, IsNil)

		err = filepath.Walk(wd, func(p string, fi os.FileInfo, err error) error {
			c.Assert(err, IsNil)

			rel, err := filepath.Rel(wd, p)
			c.Assert(err, IsNil)
			header, err := tar.FileInfoHeader(fi, "")
			c.Assert(err, IsNil)
			header.Name = rel
			c.Assert(tw.WriteHeader(header), IsNil)

			if !fi.IsDir() {
				abs, err := filepath.Abs(p)
				c.Assert(err, IsNil)
				f, err := os.Open(abs)
				c.Assert(err, IsNil)
				_, err = io.Copy(tw, f)
				c.Assert(err, IsNil)
				return f.Close()
			}

			_, err = tw.Write(nil)
			return err
		})

		w.CloseWithError(err)
		tw.Close()
	}
	return m.Repository.NewImage(parent), parent
}
