package overmount

import (
	"compress/gzip"
	"net/http"
	"os"
	"path"

	digest "github.com/opencontainers/go-digest"

	. "gopkg.in/check.v1"
)

func (m *mountSuite) TestLayerProperties(c *C) {
	layer, err := m.Repository.CreateLayer("test", nil)
	c.Assert(err, IsNil)
	_, err = m.Repository.CreateLayer("test", nil)
	c.Assert(err, Equals, ErrLayerExists)
	c.Assert(path.Base(layer.MountPath()), Equals, "test")
	c.Assert(path.Dir(layer.MountPath()), Equals, path.Join(m.Repository.baseDir, mountBase))
	c.Assert(path.Base(path.Dir(layer.Path())), Equals, "test")
	c.Assert(path.Dir(path.Dir(layer.Path())), Equals, path.Join(m.Repository.baseDir, layerBase))
	resp, err := http.Get("https://storage.googleapis.com/golang/go1.7.4.linux-amd64.tar.gz")
	c.Assert(err, IsNil)
	defer resp.Body.Close()
	d1 := layer.asset.Digest()
	c.Assert(d1, Equals, digest.Digest(emptyDigest))
	gz, err := gzip.NewReader(resp.Body)
	c.Assert(err, IsNil)
	d2, err := layer.Unpack(gz)
	c.Assert(err, IsNil)
	c.Assert(d2, Not(Equals), digest.Digest("sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"))

	dg := digest.SHA256.Digester()
	dg2, err := layer.Pack(dg.Hash())
	c.Assert(err, IsNil)
	c.Assert(dg.Digest(), Equals, d2)
	c.Assert(dg2, Equals, dg.Digest())
	c.Assert(layer.Exists(), Equals, true)
	c.Assert(layer.Remove(), IsNil)
	c.Assert(layer.Exists(), Equals, false)
}

func (m *mountSuite) TestLayerParentCommit(c *C) {
	layerCount := 10

	_, layer := m.makeImage(c, layerCount)

	for iter := layer; iter != nil; iter = iter.Parent {
		c.Assert(iter.SaveParent(), IsNil)
		c.Assert(iter.SaveParent(), IsNil) // double save should have no error
	}

	var err error

	parentID := layer.Parent.ID()
	id := layer.ID()
	m.Repository, err = NewRepository(m.Repository.baseDir, os.Getenv("VIRTUAL") != "")
	c.Assert(err, IsNil)
	layer, err = m.Repository.NewLayer(id, nil)
	c.Assert(err, IsNil)
	c.Assert(layer.RestoreParent(), IsNil)

	var count int
	for iter := layer; iter != nil; iter = iter.Parent {
		count++
	}

	c.Assert(count, Equals, layerCount)
	m.Repository, err = NewRepository(m.Repository.baseDir, os.Getenv("VIRTUAL") != "")
	c.Assert(err, IsNil)
	layer, err = m.Repository.NewLayer(id, nil)
	layer.Parent = nil
	c.Assert(layer.LoadParent(), IsNil)
	c.Assert(layer.Parent, NotNil)
	c.Assert(layer.Parent.ID(), Equals, parentID)
}

func (m *mountSuite) TestLayerConfig(c *C) {
	_, layer := m.makeImage(c, 10)
	config, err := layer.Config()
	c.Assert(config, IsNil)
	c.Assert(err, NotNil)
	c.Assert(layer.SaveConfig(&ImageConfig{Cmd: []string{"quux"}}), IsNil)
	config, err = layer.Config()
	c.Assert(err, IsNil)
	c.Assert(config.Cmd, DeepEquals, []string{"quux"})
}
