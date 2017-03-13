package overmount

import (
	"io"

	digest "github.com/opencontainers/go-digest"
)

// AssetReader is the reader representation of an on-disk asset
type AssetReader interface {
	Digest() digest.Digest
	io.ReadCloser
}

// AssetFS is a filesystem-backed asset
type AssetFS string

// AssetTar is a tar-backed asset
type AssetTar io.ReadCloser

// AssetNil performs no operations and is used for testing.
type AssetNil struct{}

// Read reads nothing from the nil reader
func (a AssetNil) Read(buf []byte) (int, error) {
	return 0, nil
}

// Close closes nothing for the nil reader
func (a AssetNil) Close() error {
	return nil
}

// Digest returns a nil digest
func (a AssetNil) Digest() digest.Digest {
	return digest.FromBytes(nil)
}
