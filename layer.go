package overmount

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// Mount the layer against any parent layers.
func (l *Layer) Mount() (*Mount, error) {
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

	return mount, nil
}
