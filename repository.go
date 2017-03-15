package overmount

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

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
func (r *Repository) NewLayer(id string, parent *Layer) *Layer {
	return &Layer{
		ID:     id,
		Parent: parent,
		//Asset:      asset,
		Repository: r,
	}
}

// MountPath gets the mount path for a given subpath, usually the layer id.
func (l *Layer) MountPath() string {
	return filepath.Join(l.Repository.BaseDir, mountBase, l.ID)
}

// Path gets the layer store path for a given subpath, usually the layer id.
func (l *Layer) Path() string {
	return filepath.Join(l.Repository.BaseDir, layerBase, l.ID)
}

// NewImage preps a set of layers to be a part of an image.
func (r *Repository) NewImage(topLayer *Layer) *Image {
	return &Image{repository: r, layer: topLayer}
}

func (r *Repository) mkdirCheckRel(path string) error {
	rel, err := filepath.Rel(r.BaseDir, path)
	if err != nil {
		return err
	}

	if strings.HasPrefix(rel, "../") {
		return errors.Wrap(ErrMountCannotProceed, "relative path falls below basedir root")
	}

	return os.MkdirAll(path, 0700)
}

// Mount mounts an image with the specified layer as its highest element.
func (i *Image) Mount() error {
	upper := i.layer.Path()
	target := i.layer.MountPath()

	layer := i.layer.Parent

	lower := ""

	for layer != nil {
		if err := i.repository.mkdirCheckRel(layer.Path()); err != nil {
			return err
		}
		if lower != "" {
			lower = layer.Path() + ":" + lower
		} else {
			lower = layer.Path()
		}
		layer = layer.Parent
	}

	for _, path := range []string{target, upper} {
		if err := i.repository.mkdirCheckRel(path); err != nil {
			return errors.Wrap(ErrMountCannotProceed, err.Error())
		}
	}

	mount, err := i.repository.NewMount(target, lower, upper)
	if err != nil {
		return err
	}

	i.mount = mount

	return mount.Open()
}

// Unmount unmounts the image.
func (i *Image) Unmount() error {
	return i.mount.Close()
}
