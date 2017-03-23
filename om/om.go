package main

import (
	"fmt"
	"os"

	"github.com/erikh/overmount"
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
			Value:  "~/.overmount",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "image",
			Usage: "Perform commands on images",
			Subcommands: []cli.Command{
				{
					Name:   "list-layers",
					Usage:  "list the layer IDs of an image",
					Action: listLayers,
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
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
