package overmount

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// Mount the layer against any parent layers.
func (l *Layer) Mount() (*Mount, error) {
	if l.Parent != nil && !l.Parent.Mounted() {
		return nil, ErrParentNotMounted
	}

	var lower string

	if l.Parent != nil {
		lower = l.Repository.MountPath(l.Parent.ID)
	} else {
		lower = l.Repository.LayerPath(l.ID)
	}

	upper := l.Repository.LayerPath(l.ID)
	target := l.Repository.MountPath(l.ID)

	for _, path := range []string{lower, upper, target} {
		t, err := filepath.Rel(l.Repository.BaseDir, path)
		if err != nil {
			return nil, err
		}

		if strings.HasPrefix(t, "../") {
			return nil, errors.Wrap(ErrMountCannotProceed, "path fell below repository root")
		}

		if err := os.MkdirAll(path, 0700); err != nil {
			return nil, errors.Wrap(ErrMountCannotProceed, err.Error())
		}
	}

	mount, err := l.Repository.NewMount(target, lower, upper)
	if err != nil {
		return nil, errors.Wrap(ErrMountCannotProceed, err.Error())
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
