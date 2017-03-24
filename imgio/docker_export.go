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

	config, err := layer.Config()
	if err != nil {
		return nil, errors.Wrap(om.ErrInvalidLayer, err.Error())
	}

	if config == nil {
		return nil, errors.Wrap(om.ErrImageCannotBeComposed, "missing image configuration")
	}

	r, w := io.Pipe()

	go func(layer *om.Layer, w *io.PipeWriter) {
		tw := tar.NewWriter(w)
		files := []*os.File{}
		defer func() {
			for _, file := range files {
				file.Close()
				os.Remove(file.Name())
			}
		}()

		layerIDs := []string{}

		for iter := layer; iter != nil; iter = iter.Parent {
			layerIDs = append(layerIDs, iter.ID())

			err := tw.WriteHeader(&tar.Header{
				Name:     iter.ID(),
				Mode:     0700,
				Typeflag: tar.TypeDir,
			})
			if err != nil {
				w.CloseWithError(errors.Wrap(om.ErrImageCannotBeComposed, "cannot add directory to tar writer"))
				return
			}

			tf, err := ioutil.TempFile("", "layer-")
			if err != nil {
				w.CloseWithError(err)
				return
			}

			files = append(files, tf)

			if _, err := iter.Pack(tf); err != nil {
				w.CloseWithError(err)
				return
			}

			if _, err := tf.Seek(0, 0); err != nil {
				w.CloseWithError(err)
				return
			}

			fi, err := tf.Stat()
			if err != nil {
				w.CloseWithError(err)
				return
			}

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
				w.CloseWithError(err)
				return
			}

			err = tw.WriteHeader(&tar.Header{
				Name:     path.Join(iter.ID(), "json"),
				Mode:     0600,
				Typeflag: tar.TypeReg,
				Size:     int64(len(content)),
			})
			if err != nil {
				w.CloseWithError(err)
				return
			}

			if _, err := tw.Write(content); err != nil {
				w.CloseWithError(err)
				return
			}

			err = tw.WriteHeader(&tar.Header{
				Name:     path.Join(iter.ID(), "layer.tar"),
				Mode:     0600,
				Typeflag: tar.TypeReg,
				Size:     fi.Size(),
			})

			if err != nil {
				w.CloseWithError(errors.Wrap(om.ErrImageCannotBeComposed, "cannot add file to tar writer"))
				return
			}

			if _, err := io.Copy(tw, tf); err != nil {
				w.CloseWithError(err)
				return
			}
		}

		content, err := json.Marshal(map[string]interface{}{})
		if err != nil {
			w.CloseWithError(err)
			return
		}

		err = tw.WriteHeader(&tar.Header{
			Name:     "repositories",
			Mode:     0600,
			Typeflag: tar.TypeReg,
			Size:     int64(len(content)),
		})
		if err != nil {
			w.CloseWithError(err)
			return
		}

		if _, err := tw.Write(content); err != nil {
			w.CloseWithError(err)
			return
		}

		content, err = json.Marshal([]map[string]interface{}{
			{
				"Config":   fmt.Sprintf("%s.json", layer.ID()),
				"RepoTags": []string{},
				"Layers":   layerIDs,
			},
		})
		if err != nil {
			w.CloseWithError(err)
			return
		}

		err = tw.WriteHeader(&tar.Header{
			Name:     "manifest.json",
			Mode:     0600,
			Typeflag: tar.TypeReg,
			Size:     int64(len(content)),
		})
		if err != nil {
			w.CloseWithError(err)
			return
		}

		if _, err := tw.Write(content); err != nil {
			w.CloseWithError(err)
			return
		}
		content, err = json.Marshal(config)
		if err != nil {
			w.CloseWithError(err)
			return
		}

		err = tw.WriteHeader(&tar.Header{
			Name:     fmt.Sprintf("%s.json", layer.ID()),
			Mode:     0600,
			Typeflag: tar.TypeReg,
			Size:     int64(len(content)),
		})
		if err != nil {
			w.CloseWithError(err)
			return
		}

		if _, err := tw.Write(content); err != nil {
			w.CloseWithError(err)
			return
		}

		if err := tw.Close(); err != nil {
			w.CloseWithError(err)
			return
		}
		w.Close()
	}(layer, w)

	return r, nil
}
