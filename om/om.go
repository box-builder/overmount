package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/docker/docker/client"
	"github.com/erikh/overmount"
	"github.com/erikh/overmount/imgio"
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
	}

	app.Commands = []cli.Command{
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
	repo, err := overmount.NewRepository(ctx.GlobalString("repo"))
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
	repo, err := overmount.NewRepository(ctx.GlobalString("repo"))
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

func importImage(ctx *cli.Context) {
	repo, err := overmount.NewRepository(ctx.GlobalString("repo"))
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

	layer, err := docker.Import(repo, reader)
	if err != nil {
		errExit(2, err)
	}

	fmt.Println(layer.ID())
}

func listLayers(ctx *cli.Context) {
	repo, err := overmount.NewRepository(ctx.GlobalString("repo"))
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
