package overmount

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/docker/docker/pkg/archive"
	digest "github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
	. "gopkg.in/check.v1"
)

func (m *mountSuite) TestAssetBasic(c *C) {
	if os.Getenv("VIRTUAL") != "" {
		c.Skip("Cannot run this test with virtual layers")
		return
	}

	tmpdir, err := ioutil.TempDir("", "")
	c.Assert(err, IsNil)

	dispatchTable := map[string]Checker{
		tmpdir:         IsNil,
		"overmount.go": NotNil,
	}

	for dir, check := range dispatchTable {
		if check == IsNil { // cheating a bit to ensure we're not removing overmount.go
			c.Assert(os.Remove(dir), IsNil)
			c.Assert(os.Symlink("/etc", dir), IsNil)
			asset, _ := NewAsset(dir, digest.SHA256.Digester(), false)
			_, err := asset.LoadDigest()
			c.Assert(errors.Cause(err), Equals, ErrInvalidAsset)
			c.Assert(os.Remove(dir), IsNil)
			asset, err = NewAsset(dir, digest.SHA256.Digester(), false)
			c.Assert(err, IsNil)
			_, err = asset.LoadDigest()
			c.Assert(errors.Cause(err), Equals, ErrInvalidAsset)
			c.Assert(os.Mkdir(dir, 0700), IsNil)
		}

		asset, err := NewAsset(dir, digest.SHA256.Digester(), false)
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

	// Testing LoadDigest
	dir := tmpdir

	asset, err := NewAsset(dir, digest.SHA256.Digester(), false)
	c.Assert(err, IsNil)
	digester := digest.SHA256.Digester()
	c.Assert(asset.Pack(digester.Hash()), IsNil)
	asset.resetDigest()
	dg, err := asset.LoadDigest()
	c.Assert(err, IsNil)
	c.Assert(dg, Equals, digester.Digest())
	c.Assert(asset.Digest(), Equals, digester.Digest())

	tf, err := ioutil.TempDir("", "overmount-temp-symlink-target-")
	c.Assert(os.RemoveAll(dir), IsNil)
	c.Assert(os.Symlink(tf, dir), IsNil)
	reader, err := archive.Tar("/go/src/github.com/box-builder/overmount", archive.Uncompressed)
	c.Assert(err, IsNil)
	c.Assert(errors.Cause(asset.Unpack(reader)), Equals, ErrInvalidAsset)
	c.Assert(errors.Cause(asset.Pack(ioutil.Discard)), Equals, ErrInvalidAsset)
}

func (m *mountSuite) TestAssetVirtual(c *C) {
	if os.Getenv("VIRTUAL") == "" {
		c.Skip("This test is for virtual layers only")
		return
	}

	dir, err := ioutil.TempDir("", "overmount-asset-test-")
	c.Assert(err, IsNil)

	layerFile := path.Join(dir, "layer.tar")

	asset, err := NewAsset(layerFile, digest.SHA256.Digester(), true)
	c.Assert(err, IsNil)
	c.Assert(asset.Path(), Equals, layerFile)
	c.Assert(asset.Digest(), Equals, emptyDigest)

	reader, err := archive.Tar("/go/src/github.com/box-builder/overmount", archive.Uncompressed)
	c.Assert(err, IsNil)
	c.Assert(asset.Unpack(reader), IsNil)
	c.Assert(asset.Digest(), Not(Equals), emptyDigest)
	fi, err := os.Stat(asset.Path())
	c.Assert(err, IsNil)
	c.Assert(fi.Size(), Not(Equals), 0)
	dg := asset.Digest()
	asset.resetDigest()
	c.Assert(asset.Digest(), Not(Equals), dg)
	_, err = asset.LoadDigest()
	c.Assert(err, IsNil)
	c.Assert(asset.Digest(), Equals, dg)

	errdir, err := ioutil.TempDir("", "overmount-asset-test-")
	c.Assert(err, IsNil)
	asset, err = NewAsset(errdir, digest.SHA256.Digester(), true)
	c.Assert(err, IsNil)
	reader, err = archive.Tar("/go/src/github.com/box-builder/overmount", archive.Uncompressed)
	c.Assert(err, IsNil)
	c.Assert(asset.Unpack(reader), NotNil)
	c.Assert(asset.Pack(ioutil.Discard), NotNil)

	tf, err := ioutil.TempFile("", "box-symlink-test-")
	c.Assert(err, IsNil)
	c.Assert(os.Remove(layerFile), IsNil)
	c.Assert(os.Symlink(tf.Name(), layerFile), IsNil)
	reader, err = archive.Tar("/go/src/github.com/box-builder/overmount", archive.Uncompressed)
	c.Assert(err, IsNil)
	c.Assert(errors.Cause(asset.Unpack(reader)), Equals, ErrInvalidAsset)
	c.Assert(errors.Cause(asset.Pack(ioutil.Discard)), Equals, ErrInvalidAsset)
}
