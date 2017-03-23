package imgio

import (
	"archive/tar"
	"io"
	"io/ioutil"
	"os"
	"path"

	om "github.com/erikh/overmount"
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

		for iter := layer; iter != nil; iter = iter.Parent {
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
		if err := tw.Close(); err != nil {
			w.CloseWithError(err)
			return
		}
		w.Close()
	}(layer, w)

	return r, nil
}
