package overmount

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// NewRepository constructs a *Repository and creates the dir in which the
// repository lives. A repository is used to hold images and mounts.
func NewRepository(baseDir string) (*Repository, error) {
	return &Repository{BaseDir: baseDir}, os.MkdirAll(baseDir, 0700)
}

// TempDir returns a temporary path within the repository
func (r *Repository) TempDir() (string, error) {
	basePath := filepath.Join(r.BaseDir, tmpdirBase)
	if err := os.MkdirAll(basePath, 0700); err != nil {
		return "", err
	}
	return ioutil.TempDir(basePath, "")
}

// MountPath gets the mount path for a given subpath, usually the layer id.
func (r *Repository) MountPath(id string) string {
	return filepath.Join(r.BaseDir, mountBase, id)
}

// LayerPath gets the layer store path for a given subpath, usually the layer id.
func (r *Repository) LayerPath(id string) string {
	return filepath.Join(r.BaseDir, layerBase, id)
}

// NewMount creates a new mount for use.
func (r *Repository) NewMount(target, lower, upper string) (*Mount, error) {
	workDir, err := r.TempDir()
	if err != nil {
		return nil, errors.Wrap(ErrMountCannotProceed, err.Error())
	}

	return &Mount{
		Target: target,
		Upper:  upper,
		Lower:  lower,
		work:   workDir,
	}, nil
}

// NewLayer prepares a new layer for work.
func (r *Repository) NewLayer(id string, parent *Layer, asset *Asset) *Layer {
	return &Layer{
		ID:         id,
		Parent:     parent,
		Asset:      asset,
		Repository: r,
	}
}

// NewImage preps a set of layers to be a part of an image.
func (r *Repository) NewImage(layers []*Layer) *Image {
	return &Image{layers: layers, mounts: []*Mount{}}
}
