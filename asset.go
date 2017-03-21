package overmount

import (
	"io"

	"github.com/docker/docker/pkg/archive"
	digest "github.com/opencontainers/go-digest"
)

const emptyDigest = digest.Digest("sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")

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

// Digest returns the digest of the last pack or unpack.
func (a *Asset) Digest() digest.Digest {
	return a.digest.Digest()
}

// Path gets the filesystem path we will be working with.
func (a *Asset) Path() string {
	return a.path
}

// Unpack from the io.Reader (must be a tar file!) and unpack to the filesystem.
// Accepts io.Reader, not *tar.Reader!
func (a *Asset) Unpack(reader io.Reader) error {
	a.resetDigest()
	if err := checkDir(a.path, ErrInvalidAsset); err != nil {
		return err
	}

	err := archive.Unpack(io.TeeReader(reader, a.digest.Hash()), a.path, &archive.TarOptions{WhiteoutFormat: archive.OverlayWhiteoutFormat})
	if err != nil {
		return err
	}

	return nil
}

// Pack a tarball from the filesystem. Accepts an io.Writer, not a
// *tar.Writer!
func (a *Asset) Pack(writer io.Writer) error {
	a.resetDigest()

	if err := checkDir(a.path, ErrInvalidAsset); err != nil {
		return err
	}

	reader, err := archive.Tar(a.path, archive.Uncompressed)
	if err != nil {
		return err
	}

	if _, err := io.Copy(writer, io.TeeReader(reader, a.digest.Hash())); err != nil {
		return err
	}

	return nil
}

// resetDigest resets the digester so it can re-calculate e.g. in a scenario
// where more than one read/write (or swapping between the two) is called.
func (a *Asset) resetDigest() {
	a.digest = digest.SHA256.Digester()
}
