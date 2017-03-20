package overmount

import (
	"io"
	"os"
	"path/filepath"

	digest "github.com/opencontainers/go-digest"
)

const (
	rootFSPath = "rootfs"
)

// NewLayer prepares a new layer for work. The ID is the directory that will be
// created in the repository; see NewRepository for more info.
func (r *Repository) NewLayer(id string, parent *Layer) (*Layer, error) {
	var err error

	layer := &Layer{
		id:         id,
		parent:     parent,
		repository: r,
	}

	if err := r.AddLayer(layer); err != nil {
		return nil, err
	}

	layer.asset, err = NewAsset(layer.Path(), digest.SHA256.Digester())
	if err != nil {
		return nil, err
	}

	return layer, nil
}

// ID returns the ID of the layer.
func (l *Layer) ID() string {
	return l.id
}

// Parent returns the parent layer.
func (l *Layer) Parent() *Layer {
	return l.parent
}

// MountPath gets the mount path for a given subpath.
func (l *Layer) MountPath() string {
	return filepath.Join(l.repository.baseDir, mountBase, l.id)
}

func (l *Layer) layerBase() string {
	return filepath.Join(l.repository.baseDir, layerBase, l.id)
}

// Path gets the layer store path for a given subpath.
func (l *Layer) Path() string {
	return filepath.Join(l.layerBase(), rootFSPath)
}

// Unpack unpacks the asset into the layer Path(). It returns the computed digest.
func (l *Layer) Unpack(reader io.Reader) (digest.Digest, error) {
	if err := l.asset.Unpack(reader); err != nil {
		return digest.Digest(""), err
	}

	return l.asset.Digest(), nil
}

// Pack archives the layer to the writer as a tar file.
func (l *Layer) Pack(writer io.Writer) (digest.Digest, error) {
	if err := l.asset.Pack(writer); err != nil {
		return digest.Digest(""), err
	}

	return l.asset.Digest(), nil
}

// Remove a layer from the filesystem and the repository.
func (l *Layer) Remove() error {
	l.repository.RemoveLayer(l)
	return os.RemoveAll(l.Path())
}
