package commands

import (
	"context"
	"fmt"
	"tbc/internal/terabox"
	"tbc/internal/util"

	"github.com/urfave/cli/v3"
)

const getCmdName = "get"

var getCmd = &cli.Command{
	Name:      getCmdName,
	Usage:     "Download file from TeraBox",
	UsageText: fmt.Sprintf("%s %s FILE", CliName, getCmdName),
	Action:    getAction,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "destination",
			Aliases:     []string{"d"},
			Usage:       "Local destination directory",
			Value:       ".",
			HideDefault: true,
		},
		&cli.IntFlag{
			Name:    "split",
			Aliases: []string{"n"},
			Usage:   "Download a file using N connections (Possible: 1-16)",
			Value:   5,
			Validator: func(arg int64) error {
				if arg < 1 || arg > 16 {
					return fmt.Errorf("--split option must be 1-16")
				}
				return nil
			},
		},
		&cli.StringFlag{
			Name:    "chunk-size",
			Aliases: []string{"s"},
			Usage:   "Split size (1k = 1024, 1M = 1048576)",
			Value:   "50M",
			Validator: func(arg string) error {
				_, err := util.ParseChunkSize(arg)
				return err
			},
		},
	},
	UseShortOptionHandling: true,
}

func getAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() < 1 {
		cli.ShowSubcommandHelp(cmd)
		return fmt.Errorf("Invalid arguments")
	}

	// Create TeraBox client
	client, err := setupClient(cmd)
	if err != nil {
		return err
	}

	chunkSize, _ := util.ParseChunkSize(cmd.String("chunk-size"))
	opts := terabox.DownloadOptions{
		DownloadChunkSize: chunkSize,
		MaxConcurrency:    cmd.Int("split"),
	}
	remotePath := cmd.Args().First()

	err = client.Download(ctx, []terabox.DownloadFile{
		{
			RemotePath: remotePath,
			LocalDir:   cmd.String("destination"),
		},
	}, nil, &opts)
	return err
}
