package overmount

// Mount an image's layers in order, unmounts all layers and returns on error
func (i *Image) Mount() error {
	for _, layer := range i.layers {
		mount, err := layer.Mount()
		if err != nil {
			return err
		}

		i.mounts = append(i.mounts, mount)
	}

	return nil
}

// Unmount an image's layers in order, returns an error on an interruption.
// Can safely be retried.
func (i *Image) Unmount() error {
	for x := len(i.mounts) - 1; x >= 0; x-- {
		if err := i.mounts[x].Close(); err != nil {
			return err
		}

		i.mounts = i.mounts[:x]
	}

	return nil
}
