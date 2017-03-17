package overmount

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

// NewRepository constructs a *Repository and creates the dir in which the
// repository lives. A repository is used to hold images and mounts.
func NewRepository(baseDir string) (*Repository, error) {
	return &Repository{
		BaseDir:   baseDir,
		Layers:    map[string]*Layer{},
		Mounts:    []*Mount{},
		editMutex: new(sync.Mutex),
	}, os.MkdirAll(baseDir, 0700)
}

// TempDir returns a temporary path within the repository
func (r *Repository) TempDir() (string, error) {
	basePath := filepath.Join(r.BaseDir, tmpdirBase)
	if err := os.MkdirAll(basePath, 0700); err != nil {
		return "", err
	}
	return ioutil.TempDir(basePath, "")
}

// NewMount creates a new mount for use. Target, lower, and upper correspond to
// specific fields in the mount stanza; read
// https://www.kernel.org/doc/Documentation/filesystems/overlayfs.txt for more
// information.
func (r *Repository) NewMount(target, lower, upper string) (*Mount, error) {
	workDir, err := r.TempDir()
	if err != nil {
		return nil, errors.Wrap(ErrMountCannotProceed, err.Error())
	}

	mount := &Mount{
		Target: target,
		Upper:  upper,
		Lower:  lower,
		work:   workDir,
	}

	if err := r.AddMount(mount); err != nil {
		return nil, err
	}

	return mount, nil
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

func (r *Repository) edit(editFunc func() error) error {
	r.editMutex.Lock()
	defer r.editMutex.Unlock()
	return editFunc()
}

// AddLayer adds a layer to the repository.
func (r *Repository) AddLayer(layer *Layer) error {
	return r.edit(func() error {
		if _, ok := r.Layers[layer.ID]; ok {
			return ErrLayerExists
		}
		r.Layers[layer.ID] = layer
		return nil
	})
}

// RemoveLayer removes a layer from the repository
func (r *Repository) RemoveLayer(layer *Layer) {
	r.edit(func() error {
		delete(r.Layers, layer.ID)
		return nil
	})
}

// AddMount adds a layer to the repository.
func (r *Repository) AddMount(mount *Mount) error {
	return r.edit(func() error {
		r.Mounts = append(r.Mounts, mount)
		return nil
	})
}

// RemoveMount removes a layer from the repository
func (r *Repository) RemoveMount(mount *Mount) {
	r.edit(func() error {
		for i, x := range r.Mounts {
			if mount.Target == x.Target && mount.Upper == x.Upper && mount.Lower == x.Lower && mount.work == x.work {
				r.Mounts = append(r.Mounts[:i], r.Mounts[i+1:]...)
			}
		}
		return nil
	})
}
