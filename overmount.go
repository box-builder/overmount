package overmount

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"golang.org/x/sys/unix"

	digest "github.com/opencontainers/go-digest"
)

var (
	// ErrParentNotMounted is returned when the parent layer is not mounted (but exists)
	ErrParentNotMounted = errors.New("parent not mounted, cannot continue")

	// ErrMountFailed returns an underlying error when the mount has failed.
	ErrMountFailed = errors.New("mount failed")

	// ErrUnmountFailed returns an underlying error when the unmount has failed.
	ErrUnmountFailed = errors.New("unmount failed")

	// ErrMountCannotProceed returns an underlying error when the mount cannot be processed.
	ErrMountCannotProceed = errors.New("mount cannot proceed")
)

const (
	tmpdirBase = "tmp"
	mountBase  = "mount"
	layerBase  = "layers"
)

// Repository is a collection of mounts/layers.
type Repository struct {
	BaseDir string
}

// Mount represents a single overlay mount
type Mount struct {
	Target string
	Upper  string
	Lower  string

	work  string
	layer *Layer
}

// Layer is the representation of a filesystem layer.
type Layer struct {
	ID         string
	Parent     *Layer
	Asset      AssetReader
	Repository *Repository

	mount *Mount
}

// AssetReader is the reader representation of an on-disk asset
type AssetReader interface {
	Digest() digest.Digest
	io.ReadCloser
}

// AssetFS is a filesystem-backed asset
type AssetFS string

// AssetTar is a tar-backed asset
type AssetTar io.ReadCloser

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
	fi, err := os.Stat(lower)
	if err != nil {
		return nil, errors.Wrap(ErrMountCannotProceed, err.Error())
	}

	if !fi.IsDir() {
		return nil, ErrMountCannotProceed
	}

	workDir, err := r.TempDir()
	if err != nil {
		return nil, errors.Wrap(ErrMountCannotProceed, err.Error())
	}

	return &Mount{
		Target: target,
		Upper:  upper,
		Lower:  lower,
		work:   workDir,
	}, err
}

// Open a mount
func (m *Mount) Open() error {
	return unix.Mount("overlay", m.Target, "overlay", 0, fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", m.Lower, m.Upper, m.work))
}

// Close a mount
func (m *Mount) Close() error {
	if err := unix.Unmount(m.Upper, 0); err != nil {
		return err
	}

	if err := os.RemoveAll(m.work); err != nil {
		return err
	}

	return nil
}

// Mounted returns true if the volume is currently mounted
func (m *Mount) Mounted() bool {
	return false
}

// NewLayer prepares a new layer for work.
func (r *Repository) NewLayer(id string, parent *Layer, asset AssetReader) *Layer {
	return &Layer{
		ID:         id,
		Parent:     parent,
		Asset:      asset,
		Repository: r,
	}
}

// Mount the layer against any parent layers.
func (l *Layer) Mount() (*Mount, error) {
	if l.Parent != nil && !l.Parent.Mounted() {
		return nil, ErrParentNotMounted
	}

	lower := l.Repository.MountPath(l.Parent.ID)
	upper := l.Repository.LayerPath(l.ID)
	target := l.Repository.MountPath(l.ID)

	mount, err := l.Repository.NewMount(target, lower, upper)
	if err != nil {
		return nil, errors.Wrap(ErrMountFailed, err.Error())
	}

	if err := mount.Open(); err != nil {
		return nil, errors.Wrap(ErrMountFailed, err.Error())
	}

	l.setMount(mount)

	return mount, nil
}

// Unmount unmounts the layer and removes the mount reference.
func (l *Layer) Unmount() error {
	if err := l.mount.Close(); err != nil {
		return errors.Wrap(ErrMountFailed, err.Error())
	}

	return nil
}

// Mounted tells us if a layer is currently mounted.
func (l *Layer) Mounted() bool {
	return l.mount != nil && l.mount.Mounted()
}

// setMount appropriately propagates the mount between the layer and mount structs.
func (l *Layer) setMount(m *Mount) {
	l.mount = m
	m.layer = l
}
