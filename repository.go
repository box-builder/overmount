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
