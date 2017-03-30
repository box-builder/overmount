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
	digest "github.com/opencontainers/go-digest"
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
	layers := []*om.Layer{}
	chainIDs := []digest.Digest{}
	diffIDs := []digest.Digest{}

	var parent digest.Digest

	// we need to walk it from the root up; so we need to reverse the list.
	for iter := layer; iter != nil; iter = iter.Parent {
		layers = append(layers, iter)
	}

	for i := len(layers) - 1; i >= 0; i-- {
		iter := layers[i]
		chainID, diffID, err := packLayer(parent, iter, tw)
		if err != nil {
			return err
		}

		chainIDs = append(chainIDs, chainID)
		diffIDs = append(diffIDs, diffID)

		if err := writeLayerConfig(chainID, parent, iter, tw); err != nil {
			return err
		}
		parent = chainID
	}

	if err := writeRepositories(tw); err != nil {
		return err
	}

	if err := writeManifest(layer, chainIDs, tw); err != nil {
		return err
	}

	if err := writeImageConfig(chainIDs[len(chainIDs)-1], diffIDs, layer, tw); err != nil {
		return err
	}

	if err := tw.Close(); err != nil {
		return err
	}

	w.Close()

	return nil
}

func packLayer(parentDigest digest.Digest, iter *om.Layer, tw *tar.Writer) (digest.Digest, digest.Digest, error) {
	tf, err := ioutil.TempFile("", "layer-")
	if err != nil {
		return "", "", err
	}

	defer func() {
		tf.Close()
		os.Remove(tf.Name())
	}()

	packDigest, err := iter.Pack(tf)
	if err != nil {
		return "", "", err
	}

	hexDigest := ""
	if parentDigest != "" {
		hexDigest = parentDigest.Hex()
	}

	chainID := digest.FromBytes([]byte(string(hexDigest) + " " + string(packDigest.Hex())))

	err = tw.WriteHeader(&tar.Header{
		Name:     chainID.Hex(),
		Mode:     0700,
		Typeflag: tar.TypeDir,
	})
	if err != nil {
		return "", "", errors.Wrap(om.ErrImageCannotBeComposed, "cannot add directory to tar writer")
	}

	if _, err := tf.Seek(0, 0); err != nil {
		return "", "", err
	}

	fi, err := tf.Stat()
	if err != nil {
		return "", "", err
	}

	err = tw.WriteHeader(&tar.Header{
		Name:     path.Join(chainID.Hex(), "layer.tar"),
		Mode:     0600,
		Typeflag: tar.TypeReg,
		Size:     fi.Size(),
	})

	if err != nil {
		return "", "", errors.Wrap(om.ErrImageCannotBeComposed, "cannot add file to tar writer")
	}

	if _, err := io.Copy(tw, tf); err != nil {
		return "", "", err
	}

	return chainID, packDigest, nil
}

func writeLayerConfig(chainID digest.Digest, parentID digest.Digest, iter *om.Layer, tw *tar.Writer) error {
	var parent string
	if parentID != "" {
		parent = parentID.Hex()
	}

	content, err := json.Marshal(map[string]interface{}{
		"id":     chainID.Hex(),
		"parent": parent,
		"config": v1.ImageConfig{},
	})
	if err != nil {
		return err
	}

	err = tw.WriteHeader(&tar.Header{
		Name:     path.Join(chainID.Hex(), "json"),
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

func writeManifest(layer *om.Layer, chainIDs []digest.Digest, tw *tar.Writer) error {
	chainIDHexs := []string{}
	for _, chainID := range chainIDs {
		chainIDHexs = append(chainIDHexs, path.Join(chainID.Hex(), "layer.tar"))
	}

	content, err := json.Marshal([]map[string]interface{}{
		{
			"Config":   fmt.Sprintf("%s.json", chainIDs[len(chainIDs)-1].Hex()),
			"RepoTags": []string{},
			"Layers":   chainIDHexs,
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

func writeImageConfig(chainID digest.Digest, diffIDs []digest.Digest, layer *om.Layer, tw *tar.Writer) error {
	config, err := layer.Config()
	if err != nil {
		return errors.Wrap(om.ErrInvalidLayer, err.Error())
	}

	if config == nil {
		return errors.Wrap(om.ErrImageCannotBeComposed, "missing image configuration")
	}

	config.RootFS.DiffIDs = diffIDs
	config.History = nil

	content, err := json.Marshal(config)
	if err != nil {
		return err
	}

	err = tw.WriteHeader(&tar.Header{
		Name:     fmt.Sprintf("%s.json", chainID.Hex()),
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
