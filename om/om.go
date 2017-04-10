package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/box-builder/overmount"
	"github.com/box-builder/overmount/imgio"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/term"
	"github.com/urfave/cli"
)

func errExit(exitCode int, err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(exitCode)
}

func main() {
	app := cli.NewApp()
	app.Description = ""
	app.Version = ""
	app.Usage = "overmount tool"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "repo, r",
			Usage:  "overmount repository to use",
			EnvVar: "OVERMOUNT_REPO",
			Value:  path.Join(os.Getenv("HOME"), ".overmount"),
		},
		cli.BoolFlag{
			Name:   "virtual",
			Usage:  "Switch on virtual repositories (keeping tar files only, no expansion of files)",
			EnvVar: "OVERMOUNT_VIRTUAL",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "layer",
			Usage: "Perform commands on layers",
			Subcommands: []cli.Command{
				{
					Name:      "tag",
					Usage:     "Tag a layer",
					ArgsUsage: "[layer ID] [tag1] [tag2] ...",
					Action:    tagLayer,
				},
				{
					Name:      "get",
					Usage:     "Get a layer by tag",
					ArgsUsage: "[tag]",
					Action:    getLayerByTag,
				},
			},
		},
		{
			Name:  "image",
			Usage: "Perform commands on images",
			Subcommands: []cli.Command{
				{
					Name:   "import",
					Usage:  "import a docker image",
					Action: importImage,
				},
				{
					Name:   "export",
					Usage:  "export a docker image to stdout",
					Action: exportImage,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "type",
							Value: "docker",
							Usage: "Set the type of image to export [docker|oci]",
						},
					},
				},
				{
					Name:   "list-layers",
					Usage:  "list the layer IDs of an image",
					Action: listLayers,
				},
				{
					Name:   "mount",
					Usage:  "perform an overlay mount on the image ID w/ children",
					Action: mountImage,
				},
				{
					Name:   "unmount",
					Usage:  "Unmount a previously-mounted layer",
					Action: unmountImage,
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func constructImage(ctx *cli.Context) (*overmount.Image, *overmount.Layer) {
	repo, err := overmount.NewRepository(ctx.GlobalString("repo"), ctx.GlobalBool("virtual"))
	if err != nil {
		errExit(2, err)
	}

	if len(ctx.Args()) != 1 {
		errExit(2, errors.New("please supply a docker image name (only one)"))
	}

	layer, err := repo.NewLayer(ctx.Args()[0], nil)
	if err != nil {
		errExit(2, err)
	} else if !layer.Exists() {
		errExit(2, errors.New("Layer does not exist"))
	}

	if err := layer.RestoreParent(); err != nil {
		errExit(2, err)
	}

	return repo.NewImage(layer), layer
}

func unmountImage(ctx *cli.Context) {
	image, layer := constructImage(ctx)
	if err := image.Unmount(); err != nil {
		errExit(2, err)
	}

	fmt.Println(layer.MountPath())
}

func mountImage(ctx *cli.Context) {
	repo, err := overmount.NewRepository(ctx.GlobalString("repo"), ctx.GlobalBool("virtual"))
	if err != nil {
		errExit(2, err)
	}

	if len(ctx.Args()) != 1 {
		errExit(2, errors.New("please supply a docker image name (only one)"))
	}

	layer, err := repo.NewLayer(ctx.Args()[0], nil)
	if err != nil {
		errExit(2, err)
	} else if !layer.Exists() {
		errExit(2, errors.New("Layer does not exist"))
	}

	if err := layer.RestoreParent(); err != nil {
		errExit(2, err)
	}

	if err := repo.NewImage(layer).Mount(); err != nil {
		errExit(2, err)
	}

	fmt.Println(layer.MountPath())
}

func exportImage(ctx *cli.Context) {
	repo, err := overmount.NewRepository(ctx.GlobalString("repo"), ctx.GlobalBool("virtual"))
	if err != nil {
		errExit(2, err)
	}

	if len(ctx.Args()) != 1 {
		errExit(2, errors.New("please supply a docker image name (only one)"))
	}

	if term.IsTerminal(os.Stdout.Fd()) {
		errExit(2, errors.New("Cannot copy to a terminal"))
	}

	layer, err := repo.NewLayer(ctx.Args()[0], nil)
	if err != nil {
		errExit(2, err)
	}

	if err := layer.RestoreParent(); err != nil {
		errExit(2, err)
	}

	var exporter overmount.Exporter

	switch ctx.String("type") {
	case "docker":
		exporter, err = imgio.NewDocker(nil)
		if err != nil {
			errExit(2, err)
		}
	case "oci":
		exporter = imgio.NewOCI()
	}

	reader, err := exporter.Export(repo, layer, []string{})
	if err != nil {
		errExit(2, err)
	}

	bytes, err := io.Copy(os.Stdout, reader)
	if err != nil {
		errExit(2, err)
	}

	fmt.Fprintf(os.Stderr, "Wrote %d bytes to stdout\n", bytes)
}

func importImage(ctx *cli.Context) {
	repo, err := overmount.NewRepository(ctx.GlobalString("repo"), ctx.GlobalBool("virtual"))
	if err != nil {
		errExit(2, err)
	}

	if len(ctx.Args()) != 1 {
		errExit(2, errors.New("please supply a docker image name (only one)"))
	}

	client, err := client.NewEnvClient()
	if err != nil {
		errExit(2, err)
	}

	docker, err := imgio.NewDocker(client)
	if err != nil {
		errExit(2, err)
	}

	reader, err := client.ImageSave(context.Background(), []string{ctx.Args()[0]})
	if err != nil {
		errExit(2, err)
	}

	layers, err := docker.Import(repo, reader)
	if err != nil {
		errExit(2, err)
	}

	for _, layer := range layers {
		fmt.Println(layer.ID())
	}
}

func listLayers(ctx *cli.Context) {
	repo, err := overmount.NewRepository(ctx.GlobalString("repo"), ctx.GlobalBool("virtual"))
	if err != nil {
		errExit(2, err)
	}

	for _, layerID := range ctx.Args() {
		layer, err := repo.NewLayer(layerID, nil)
		if err != nil {
			errExit(2, err)
		}

		if err := layer.RestoreParent(); err != nil {
			errExit(2, err)
		}

		var depth int
		for iter := layer; iter != nil; iter = iter.Parent {
			fmt.Printf("(depth %d): %v\n", depth, iter.ID())
			depth++
		}
	}
}

func tagLayer(ctx *cli.Context) {
	repo, err := overmount.NewRepository(ctx.GlobalString("repo"), ctx.GlobalBool("virtual"))
	if err != nil {
		errExit(2, err)
	}

	if len(ctx.Args()) < 2 {
		errExit(2, errors.New("invalid arguments"))
	}

	tags := ctx.Args()[1:]
	layerID := ctx.Args()[0]

	layer, err := repo.NewLayer(layerID, nil)
	if err != nil {
		errExit(2, err)
	}

	if !layer.Exists() {
		errExit(2, errors.New("layer does not exist"))
	}

	for _, tag := range tags {
		fmt.Printf("Tagging %v with tag %q\n", layer.ID(), tag)
		if err := repo.AddTag(tag, layer); err != nil {
			errExit(2, err)
		}
	}
}

func getLayerByTag(ctx *cli.Context) {
	repo, err := overmount.NewRepository(ctx.GlobalString("repo"), ctx.GlobalBool("virtual"))
	if err != nil {
		errExit(2, err)
	}

	if len(ctx.Args()) != 1 {
		errExit(2, errors.New("invalid arguments"))
	}

	tag := ctx.Args()[0]

	layer, err := repo.GetTag(tag)
	if err != nil {
		errExit(2, err)
	}

	fmt.Println(layer.ID())
}
