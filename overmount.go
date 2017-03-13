package overmount

import (
	"github.com/pkg/errors"
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

	// ErrImageCannotBeComposed is returned when an image (a set of layers) fails validation.
	ErrImageCannotBeComposed = errors.New("image cannot be composed")
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

	work    string
	layer   *Layer
	mounted bool
}

// Layer is the representation of a filesystem layer.
type Layer struct {
	ID         string
	Parent     *Layer
	Asset      AssetReader
	Repository *Repository

	mount *Mount
}

// Image is the representation of a set of sequential layers to be mounted.
type Image []*Layer
