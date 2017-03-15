package overmount

import (
	"fmt"
	"os"

	"github.com/pkg/errors"

	"golang.org/x/sys/unix"
)

func (m *Mount) makeMountOptions() (string, error) {
	if m.Lower == "" {
		return "", errors.Wrap(ErrMountCannotProceed, "No lower dir specified (only one layer?)")
	}

	return fmt.Sprintf("upperdir=%s,lowerdir=%s,workdir=%s", m.Upper, m.Lower, m.work), nil
}

// Open a mount
func (m *Mount) Open() error {
	opts, err := m.makeMountOptions()
	if err != nil {
		return err
	}

	if err := unix.Mount("overlay", m.Target, "overlay", 0, opts); err != nil {
		return err
	}

	m.mounted = true
	return nil
}

// Close a mount
func (m *Mount) Close() error {
	if err := unix.Unmount(m.Target, 0); err != nil {
		return err
	}

	if err := os.RemoveAll(m.work); err != nil {
		return err
	}

	m.mounted = false
	return nil
}

// Mounted returns true if the volume is currently mounted
func (m *Mount) Mounted() bool {
	return m.mounted
}
