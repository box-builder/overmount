package overmount

import (
	"os"

	"github.com/pkg/errors"
)

// checkDir validates the directory is not a symlink and exists.
func checkDir(path string, wrapErr error) error {
	fi, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(path, 0700); err != nil {
				return errors.Wrapf(wrapErr, "unable to mkdir: %v", err.Error())
			}
			return nil
		}
		return err
	}

	if !fi.IsDir() {
		return errors.Wrap(wrapErr, "not a directory")
	}

	if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
		// here we attempt to remove a whole class of potential bugs.
		return errors.Wrap(wrapErr, "cannot operate on a symlink")
	}

	return nil
}
