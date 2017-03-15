package overmount

import (
	"io"
	"os"

	"github.com/docker/docker/pkg/archive"
	digest "github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
)

// Asset is the reader representation of an on-disk asset
type Asset struct {
	path   string
	digest digest.Digester
}

// NewAsset constructs a new *Asset that operates on path `path`.
func NewAsset(path string, digest digest.Digester) (*Asset, error) {
	a := &Asset{
		path:   path,
		digest: digest,
	}

	return a, nil
}

// Mkdir attempts to make the layer directory
func (a *Asset) checkDir() error {
	fi, err := os.Lstat(a.Path())
	if err != nil {
		return err
	}

	if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
		// here we attempt to remove a whole class of potential bugs.
		return errors.Wrap(ErrInvalidAsset, "cannot operate on a symlink")
	}

	return nil
}

// Digest returns the digest; Read() or Write() must be called first!
func (a *Asset) Digest() digest.Digest {
	return a.digest.Digest()
}

// Path gets the filesystem path we will be working with.
func (a *Asset) Path() string {
	return a.path
}

// Read from the *tar.Reader and unpack on to the filesystem.
func (a *Asset) Read(reader io.Reader) error {
	if err := a.checkDir(); err != nil {
		return err
	}

	err := archive.Unpack(io.TeeReader(reader, a.digest.Hash()), a.path, &archive.TarOptions{WhiteoutFormat: archive.OverlayWhiteoutFormat})
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

	if _, err := io.Copy(writer, io.TeeReader(reader, a.digest.Hash())); err != nil {
		return err
	}

	return nil
}
