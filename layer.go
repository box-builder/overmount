package overmount

import (
	"io"
	"path/filepath"

	digest "github.com/opencontainers/go-digest"
)

// NewLayer prepares a new layer for work. The ID is the directory that will be
// created in the repository; see NewRepository for more info.
func (r *Repository) NewLayer(id string, parent *Layer) (*Layer, error) {
	var err error

	layer := &Layer{
		ID:         id,
		Parent:     parent,
		Repository: r,
	}

	layer.Asset, err = NewAsset(layer.Path(), digest.SHA256.Digester())
	if err != nil {
		return nil, err
	}

	return layer, nil
}

// MountPath gets the mount path for a given subpath.
func (l *Layer) MountPath() string {
	return filepath.Join(l.Repository.BaseDir, mountBase, l.ID)
}

// Path gets the layer store path for a given subpath.
func (l *Layer) Path() string {
	return filepath.Join(l.Repository.BaseDir, layerBase, l.ID)
}

// Unpack unpacks the asset into the layer Path(). It returns the computed digest.
func (l *Layer) Unpack(reader io.Reader) (digest.Digest, error) {
	if err := l.Asset.Read(reader); err != nil {
		return digest.Digest(""), err
	}

	return l.Asset.Digest(), nil
}
