package cmd

import (
	"context"
	"fmt"
	"tbc/internal/runner"
	"tbc/internal/terabox"
	"tbc/internal/util"

	"github.com/urfave/cli/v3"
)

const getCmdName = "get"

var getCmd = &cli.Command{
	Name:      getCmdName,
	Usage:     "Download files or directories from TeraBox",
	UsageText: fmt.Sprintf("%s %s FILE...", CliName, getCmdName),
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
			Usage:   "Split size (1M = 1048576, 1G = 1073741824)",
			Value:   "50M",
			Validator: func(arg string) error {
				size, err := util.ParseChunkSize(arg)
				if size < 1024*1024 {
					return fmt.Errorf("Split size must be greater than 1M")
				}
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

	cookie, err := util.GetCookie(cmd.Root().String(CookieFileOptName))
	if err != nil {
		return err
	}

	// Create TeraBox client
	client, err := terabox.NewClient(cookie)
	if err != nil {
		return err
	}

	destDir := cmd.String("destination")
	chunkSize, _ := util.ParseChunkSize(cmd.String("chunk-size"))
	opts := terabox.DownloadOptions{
		DownloadChunkSize: chunkSize,
		MaxConcurrency:    cmd.Int("split"),
	}

	return runner.Download(ctx, client, cmd.Args().Slice(), destDir, &opts)
}
