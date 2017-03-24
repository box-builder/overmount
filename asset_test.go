package overmount

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/docker/docker/pkg/archive"
	digest "github.com/opencontainers/go-digest"
	. "gopkg.in/check.v1"
)

func (m *mountSuite) TestAssetBasic(c *C) {
	tmpdir, err := ioutil.TempDir("", "")
	c.Assert(err, IsNil)

	dispatchTable := map[string]Checker{
		tmpdir:         IsNil,
		"overmount.go": NotNil,
	}

	for dir, check := range dispatchTable {
		asset, err := NewAsset(dir, digest.SHA256.Digester())
		c.Assert(err, IsNil)
		c.Assert(asset.Path(), Equals, dir)
		c.Assert(asset.Digest(), Equals, emptyDigest)

		reader, err := archive.Tar("/go/src/github.com/box-builder/overmount", archive.Uncompressed)
		c.Assert(err, IsNil)
		c.Assert(asset.Unpack(reader), check)
		dg := asset.Digest()
		_, err = os.Stat(path.Join(asset.Path(), "overmount.go"))
		c.Assert(err, check)
		digester := digest.SHA256.Digester()
		c.Assert(asset.Pack(digester.Hash()), check)
		c.Assert(dg, Equals, digester.Digest())
		c.Assert(dg, Equals, asset.Digest())
	}
}
