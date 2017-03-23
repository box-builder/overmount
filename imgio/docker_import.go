package imgio

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/docker/docker/pkg/archive"
	digest "github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"

	om "github.com/erikh/overmount"
)

type unpackedImage struct {
	tempdir        string
	image          *v1.Image
	layers         map[string]*om.Layer
	layerFileMap   map[string]string
	layerParentMap map[string]string
}

// Import takes a tar represented as an io.Reader, and converts and unpacks
// it into the overmount repository.  Returns the top-most layer and any
// error.
func (d *Docker) Import(r *om.Repository, reader io.ReadCloser) (*om.Layer, error) {
	tempdir, err := ioutil.TempDir("", "overmount-unpack-")
	if err != nil {
		return nil, err
	}

	defer os.RemoveAll(tempdir)

	if err := archive.Untar(reader, tempdir, &archive.TarOptions{}); err != nil {
		return nil, err
	}

	reader.Close()

	up, err := d.unpackLayers(r, tempdir)
	if err != nil {
		return nil, err
	}

	return d.constructImage(up)
}

func (d *Docker) constructImage(up *unpackedImage) (*om.Layer, error) {
	digestMap := map[digest.Digest]*om.Layer{}

	for layerID, filename := range up.layerFileMap {
		layer, ok := up.layers[layerID]
		if !ok {
			return nil, errors.Errorf("invalid layer id %v", layerID)
		}

		f, err := os.Open(filename)
		if err != nil {
			return nil, err
		}

		defer f.Close()
		layer.Parent = up.layers[up.layerParentMap[layerID]]

		var dg digest.Digest

		dg, err = layer.Unpack(f)
		if err == nil {
			if err := layer.SaveParent(); err != nil {
				return nil, err
			}
		} else if !os.IsExist(err) {
			return nil, err
		}

		digestMap[dg] = layer
	}

	topLayer := digest.Digest(up.image.RootFS.DiffIDs[len(up.image.RootFS.DiffIDs)-1])
	top, ok := digestMap[topLayer]
	if !ok {
		return nil, errors.New("top layer doesn't appear to exist")
	}

	return top, top.SaveConfig(&up.image.Config)
}

func (d *Docker) unpackLayers(r *om.Repository, tempdir string) (*unpackedImage, error) {
	up := &unpackedImage{
		tempdir:        tempdir,
		layerFileMap:   map[string]string{},
		layerParentMap: map[string]string{},
		layers:         map[string]*om.Layer{},
		image:          &v1.Image{},
	}

	err := filepath.Walk(tempdir, func(p string, fi os.FileInfo, err error) error {
		if path.Base(p) == "layer.tar" {
			f, err := os.Open(filepath.Join(path.Dir(p), "json"))
			if err != nil {
				return err
			}

			layerJSON := map[string]interface{}{}

			if err := json.NewDecoder(f).Decode(&layerJSON); err != nil {
				f.Close()
				return err
			}
			f.Close()

			if _, ok := layerJSON["id"]; !ok {
				return errors.New("missing layer id")
			}

			layerID, ok := layerJSON["id"].(string)
			if !ok {
				return errors.New("invalid layer id")
			}

			up.layerFileMap[layerID] = p

			if _, ok := layerJSON["parent"]; ok {
				up.layerParentMap[layerID], ok = layerJSON["parent"].(string)
				if !ok {
					return errors.New("invalid parent ID")
				}
			}

			layer, err := r.CreateLayer(layerID, nil)
			if err != nil {
				return err
			}

			up.layers[layerID] = layer
		} else if path.Ext(p) == ".json" && path.Base(p) != "manifest.json" {
			content, err := ioutil.ReadFile(p)
			if err != nil {
				return err
			}

			if err := json.Unmarshal(content, up.image); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return up, nil
}
