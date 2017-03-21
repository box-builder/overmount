package overmount

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	digest "github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
)

const (
	rootFSPath  = "rootfs"
	parentsPath = "parents.json"
)

// CreateLayer prepares a new layer for work and creates it in the repository.
func (r *Repository) CreateLayer(id string, parent *Layer) (*Layer, error) {
	return r.newLayer(id, parent, true)
}

// NewLayer prepares a new layer for work but DOES NOT add it to the
// repository. The ID is the directory that will be created in the repository;
// see NewRepository for more info.
func (r *Repository) NewLayer(id string, parent *Layer) (*Layer, error) {
	return r.newLayer(id, parent, false)
}

func (r *Repository) newLayer(id string, parent *Layer, create bool) (*Layer, error) {
	var err error

	layer := &Layer{
		id:         id,
		parent:     parent,
		repository: r,
	}

	if create {
		if err := layer.Create(); err != nil {
			return layer, err // return the layer here (document later) in case they need to clean it up.
		}
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

// Create creates the layer and makes it available for use, if possible.
// Otherwise, it returns an error.
func (l *Layer) Create() error {
	return checkDir(l.layerBase(), ErrInvalidLayer)
}

func (l *Layer) layerBase() string {
	return filepath.Join(l.repository.baseDir, layerBase, l.id)
}

// Path gets the layer store path for a given subpath.
func (l *Layer) Path() string {
	return filepath.Join(l.layerBase(), rootFSPath)
}

func (l *Layer) parentsPath() string {
	return filepath.Join(l.layerBase(), parentsPath)
}

// SaveParent will silently only save the
func (l *Layer) SaveParent() error {
	if l.parent == nil {
		return nil
	}

	fi, err := os.Stat(l.parentsPath())
	if err != nil {
		if os.IsNotExist(err) {
			return l.OverwriteParent()
		}
		return err
	} else if !fi.Mode().IsRegular() {
		return errors.Wrap(ErrInvalidLayer, "parent configuration is invalid")
	}

	return nil
}

// OverwriteParent overwrites the parent setting for this layer.
func (l *Layer) OverwriteParent() error {
	if l.parent == nil {
		return nil
	}

	return ioutil.WriteFile(l.parentsPath(), []byte(l.parent.ID()), 0600)
}

// LoadParent loads only the parent for this specific instance. See
// RestoreParent for restoring the whole chain.
func (l *Layer) LoadParent() error {
	id, err := ioutil.ReadFile(l.parentsPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if len(id) == 0 {
		return nil
	}

	parent, err := l.repository.NewLayer(string(id), nil)
	if err != nil {
		return err
	}

	fi, err := os.Stat(parent.layerBase())
	if err != nil || !fi.IsDir() {
		return errors.Wrap(ErrInvalidLayer, parent.layerBase())
	}

	l.parent = parent

	return nil
}

// RestoreParent reads any parent file and sets the layer accordingly. It does this recursively.
func (l *Layer) RestoreParent() error {
	if err := l.LoadParent(); err != nil {
		return err
	}

	if l.parent != nil {
		return l.parent.RestoreParent()
	}

	return nil
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
