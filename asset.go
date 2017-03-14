package overmount

import (
	"errors"
	"io"
	"os"

	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/chrootarchive"
	digest "github.com/opencontainers/go-digest"
)

// Asset is the reader representation of an on-disk asset
type Asset struct {
	path   string
	digest digest.Digest
}

// NewAsset constructs a new *Asset that operates on path `path`.
func NewAsset(path string, digest digest.Digest) (*Asset, error) {
	fi, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}

	if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
		// here we attempt to remove a whole class of potential bugs.
		return nil, errors.New("cannot operate on a symlink")
	}

	a := &Asset{
		path:   path,
		digest: digest,
	}

	return a, nil
}

// Path gets the filesystem path we will be working with.
func (a *Asset) Path() string {
	return a.path
}

// Read from the *tar.Reader and unpack on to the filesystem.
func (a *Asset) Read(reader io.Reader) error {
	_, err := chrootarchive.ApplyLayer(a.path, io.TeeReader(reader, a.digest))
	if err != nil {
		return err
	}

	return nil
}

// Write a tarball from the filesystem.
func (a *Asset) Write(writer io.Writer) error {
	reader, err := archive.Tar(a.path, archive.Uncompressed)
	if err != nil {
		return err
	}

	if _, err := io.Copy(writer, io.TeeReader(reader, a.digest)); err != nil {
		return err
	}

	return nil
}
