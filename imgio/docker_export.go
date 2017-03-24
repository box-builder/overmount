package imgio

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	om "github.com/box-builder/overmount"
	"github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
)

// Export produces a tar represented as an io.ReadCloser from the Layer provided.
func (d *Docker) Export(layer *om.Layer) (io.ReadCloser, error) {
	if !layer.Exists() {
		return nil, errors.Wrap(om.ErrInvalidLayer, "layer does not exist")
	}

	r, w := io.Pipe()
	go writeTar(layer, w)

	return r, nil
}

func writeTar(layer *om.Layer, w *io.PipeWriter) (retErr error) {
	defer func() {
		if retErr == nil {
			w.Close()
		} else {
			w.CloseWithError(retErr)
		}
	}()

	tw := tar.NewWriter(w)
	layerIDs := []string{}

	for iter := layer; iter != nil; iter = iter.Parent {
		layerIDs = append(layerIDs, iter.ID())

		if err := writeIDDir(iter, tw); err != nil {
			return err
		}

		if err := packLayer(iter, tw); err != nil {
			return err
		}

		if err := writeLayerConfig(iter, tw); err != nil {
			return err
		}
	}

	if err := writeRepositories(tw); err != nil {
		return err
	}

	if err := writeManifest(layer, layerIDs, tw); err != nil {
		return err
	}

	if err := writeImageConfig(layer, tw); err != nil {
		return err
	}

	if err := tw.Close(); err != nil {
		return err
	}

	w.Close()

	return nil
}

func writeIDDir(iter *om.Layer, tw *tar.Writer) error {
	err := tw.WriteHeader(&tar.Header{
		Name:     iter.ID(),
		Mode:     0700,
		Typeflag: tar.TypeDir,
	})
	if err != nil {
		return errors.Wrap(om.ErrImageCannotBeComposed, "cannot add directory to tar writer")
	}

	return nil
}

func packLayer(iter *om.Layer, tw *tar.Writer) error {
	tf, err := ioutil.TempFile("", "layer-")
	if err != nil {
		return err
	}

	defer func() {
		tf.Close()
		os.Remove(tf.Name())
	}()

	if _, err := iter.Pack(tf); err != nil {
		return err
	}

	if _, err := tf.Seek(0, 0); err != nil {
		return err
	}

	fi, err := tf.Stat()
	if err != nil {
		return err
	}

	err = tw.WriteHeader(&tar.Header{
		Name:     path.Join(iter.ID(), "layer.tar"),
		Mode:     0600,
		Typeflag: tar.TypeReg,
		Size:     fi.Size(),
	})

	if err != nil {
		return errors.Wrap(om.ErrImageCannotBeComposed, "cannot add file to tar writer")
	}

	if _, err := io.Copy(tw, tf); err != nil {
		return err
	}

	return nil
}

func writeLayerConfig(iter *om.Layer, tw *tar.Writer) error {
	var parent string
	if iter.Parent != nil {
		parent = iter.Parent.ID()
	}

	content, err := json.Marshal(map[string]interface{}{
		"id":     iter.ID(),
		"parent": parent,
		"config": v1.ImageConfig{},
	})
	if err != nil {
		return err
	}

	err = tw.WriteHeader(&tar.Header{
		Name:     path.Join(iter.ID(), "json"),
		Mode:     0600,
		Typeflag: tar.TypeReg,
		Size:     int64(len(content)),
	})
	if err != nil {
		return err
	}

	if _, err := tw.Write(content); err != nil {
		return err
	}

	return nil
}

func writeRepositories(tw *tar.Writer) error {
	content, err := json.Marshal(map[string]interface{}{})
	if err != nil {
		return err
	}

	err = tw.WriteHeader(&tar.Header{
		Name:     "repositories",
		Mode:     0600,
		Typeflag: tar.TypeReg,
		Size:     int64(len(content)),
	})
	if err != nil {
		return err
	}

	if _, err := tw.Write(content); err != nil {
		return err
	}

	return nil
}

func writeManifest(layer *om.Layer, layerIDs []string, tw *tar.Writer) error {
	content, err := json.Marshal([]map[string]interface{}{
		{
			"Config":   fmt.Sprintf("%s.json", layer.ID()),
			"RepoTags": []string{},
			"Layers":   layerIDs,
		},
	})
	if err != nil {
		return err
	}

	err = tw.WriteHeader(&tar.Header{
		Name:     "manifest.json",
		Mode:     0600,
		Typeflag: tar.TypeReg,
		Size:     int64(len(content)),
	})
	if err != nil {
		return err
	}

	if _, err := tw.Write(content); err != nil {
		return err
	}

	return nil
}

func writeImageConfig(layer *om.Layer, tw *tar.Writer) error {
	config, err := layer.Config()
	if err != nil {
		return errors.Wrap(om.ErrInvalidLayer, err.Error())
	}

	if config == nil {
		return errors.Wrap(om.ErrImageCannotBeComposed, "missing image configuration")
	}

	content, err := json.Marshal(config)
	if err != nil {
		return err
	}

	err = tw.WriteHeader(&tar.Header{
		Name:     fmt.Sprintf("%s.json", layer.ID()),
		Mode:     0600,
		Typeflag: tar.TypeReg,
		Size:     int64(len(content)),
	})
	if err != nil {
		return err
	}

	if _, err := tw.Write(content); err != nil {
		return err
	}

	return nil
}
