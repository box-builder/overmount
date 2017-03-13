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

func (m *mountSuite) TestMount(c *C) {
	l := m.Repository.NewLayer("one", nil, AssetNil(struct{}{}))
	mount, err := l.Mount()
	c.Assert(err, IsNil)
	c.Assert(mount.Mounted(), Equals, true)
}
