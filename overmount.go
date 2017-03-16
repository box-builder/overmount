// Package overmount - mount tars in an overlay filesystem
//
// overmount is intended to mount docker images, or work with similar
// functionality to achieve a series of layered filesystems which can be composed
// into an image.
//
// See the examples/ directory for examples of how to use the API.
//
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

	// ErrInvalidAsset is returned when the asset cannot be used.
	ErrInvalidAsset = errors.New("invalid asset")
)

const (
	tmpdirBase = "tmp"
	mountBase  = "mount"
	layerBase  = "layers"
)

// Repository is a collection of mounts/layers. Repositories have a base path
// and a collection of layers and mounts. Overlay work directories are stored
// in `tmp`.
//
// In summary:
//
//     basedir/
//        layers/
//          layer-id/
//          top-layer/
//        tmp/
//          some-random-workdir/
//        mount/
//          another-layer-id/
//          top-layer/
//
// Repositories can hold any number of mounts and layers. They do not
// necessarily need to be related.
type Repository struct {
	BaseDir string
}

// Mount represents a single overlay mount. The lower value is computed from
// the parent layer of the layer provided to the NewMount call. The target and
// upper dirs are computed from the passed layer.
type Mount struct {
	Target string
	Upper  string
	Lower  string

	work    string
	mounted bool
}

// Layer is the representation of a filesystem layer. Layers are organized in a
// reverse linked-list from topmost layer to the root layer. In an
// (*Image).Mount() scenario, the layers are mounted from the bottom up to
// culminate in a mount path that represents the top-most layer merged with all
// the lower layers.
//
// See https://www.kernel.org/doc/Documentation/filesystems/overlayfs.txt for
// more information on mount flags.
type Layer struct {
	ID         string
	Parent     *Layer
	Asset      *Asset
	Repository *Repository
}

// Image is the representation of a set of sequential layers to be mounted.
type Image struct {
	repository *Repository
	layer      *Layer
	mount      *Mount
}
