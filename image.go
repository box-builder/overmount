package overmount

import "github.com/pkg/errors"

// NewImage preps a set of layers to be a part of an image.
func (r *Repository) NewImage(topLayer *Layer) *Image {
	return &Image{repository: r, layer: topLayer}
}

// Mount mounts an image with the specified layer as its highest element.
func (i *Image) Mount() error {
	upper := i.layer.Path()
	target := i.layer.MountPath()

	layer := i.layer.Parent

	lower := ""

	for layer != nil {
		if err := i.repository.mkdirCheckRel(layer.Path()); err != nil {
			return err
		}
		if lower != "" {
			lower = layer.Path() + ":" + lower
		} else {
			lower = layer.Path()
		}
		layer = layer.Parent
	}

	for _, path := range []string{target, upper} {
		if err := i.repository.mkdirCheckRel(path); err != nil {
			return errors.Wrap(ErrMountCannotProceed, err.Error())
		}
	}

	mount, err := i.repository.NewMount(target, lower, upper)
	if err != nil {
		return err
	}

	i.mount = mount

	return mount.Open()
}

// Unmount unmounts the image.
func (i *Image) Unmount() error {
	return i.mount.Close()
}
