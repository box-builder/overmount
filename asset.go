package overmount

import (
	"io"
	"os"

	"github.com/docker/docker/pkg/archive"
	digest "github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
)

// Asset is the representation of an on-disk asset. Assets usually are a pair
// of (path, tar) where one direction is applied; f.e., you can copy from the
// tar to the dir, or the dir to the tar using the Read and Write calls.
type Asset struct {
	path   string
	digest digest.Digester
}

// NewAsset constructs a new *Asset that operates on path `path`. A digester
// must be provided. Typically this is a `digest.SHA256.Digester()` but can be
// any algorithm that opencontainers/go-digest supports.
func NewAsset(path string, digest digest.Digester) (*Asset, error) {
	a := &Asset{
		path:   path,
		digest: digest,
	}

	return a, nil
}

// checkDir validates the directory is not a symlink and exists.
func (a *Asset) checkDir() error {
	fi, err := os.Lstat(a.Path())
	if err != nil {
		return err
	}

	if !fi.IsDir() {
		return errors.Wrap(ErrInvalidAsset, "not a directory")
	}

	if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
		// here we attempt to remove a whole class of potential bugs.
		return errors.Wrap(ErrInvalidAsset, "cannot operate on a symlink")
	}

	return nil
}

// Digest returns the digest of the read/write; Read() or Write() must be
// called first!
func (a *Asset) Digest() digest.Digest {
	return a.digest.Digest()
}

// Path gets the filesystem path we will be working with.
func (a *Asset) Path() string {
	return a.path
}

// Read from the io.Reader (must be a tar file!) and unpack to the filesystem.
// Accepts io.Reader, not *tar.Reader!
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

// Write a tarball from the filesystem. Accepts an io.Writer, not a
// *tar.Writer!
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
