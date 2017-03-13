package overmount

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	. "testing"

	. "gopkg.in/check.v1"
)

type mountSuite struct {
	Repository *Repository
}

var _ = Suite(&mountSuite{})

func TestOvermount(t *T) {
	TestingT(t)
}

func (m *mountSuite) SetUpTest(c *C) {
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		panic(err)
	}
	repo, err := NewRepository(tmpdir)
	if err != nil {
		panic(err)
	}

	m.Repository = repo
}

func (m *mountSuite) TearDownTest(c *C) {
	os.RemoveAll(m.Repository.BaseDir)
}

func (m *mountSuite) TestRepositoryTempDir(c *C) {
	t, err := m.Repository.TempDir()
	c.Assert(err, IsNil)

	p, err := filepath.Rel(m.Repository.BaseDir, t)
	c.Assert(err, IsNil)
	first, _ := path.Split(p)
	c.Assert(err, IsNil)
	c.Assert(first, Equals, tmpdirBase+"/")
}

func (m *mountSuite) TestBasicLayerMount(c *C) {
	l := m.Repository.NewLayer("one", nil, AssetNil(struct{}{}))
	mount, err := l.Mount()
	c.Assert(err, IsNil)
	c.Assert(mount.Lower, Equals, mount.Upper)
	c.Assert(mount.Mounted(), Equals, true)
	c.Assert(l.ID, Equals, "one")
	c.Assert(l.Mounted(), Equals, true)
	l2 := m.Repository.NewLayer("two", l, AssetNil(struct{}{}))
	c.Assert(l2.Parent, DeepEquals, l)
	mount2, err := l2.Mount()
	c.Assert(err, IsNil)
	c.Assert(mount2.Mounted(), Equals, true)
	c.Assert(l2.Mounted(), Equals, true)
	c.Assert(mount2.Lower, Equals, mount.Target)
	c.Assert(mount2.Lower, Not(Equals), mount2.Upper)
	c.Assert(mount2.Target, Not(Equals), mount.Target)
	c.Assert(mount2.Target, Not(Equals), mount2.Upper)
	c.Assert(ioutil.WriteFile(path.Join(mount2.Target, "test"), nil, 0644), IsNil)
	_, err = os.Stat(path.Join(mount2.Upper, "test"))
	c.Assert(err, IsNil)
	_, err = os.Stat(path.Join(mount2.Lower, "test"))
	c.Assert(err, NotNil)
	_, err = os.Stat(path.Join(mount.Target, "test"))
	c.Assert(err, NotNil)
	_, err = os.Stat(path.Join(mount.Upper, "test"))
	c.Assert(err, NotNil)
}
