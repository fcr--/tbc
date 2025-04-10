package cmd

import (
	"context"
	"fmt"
	"os"
	"path"
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

	return recursiveDownload(ctx, client, cmd.Args().Slice(), destDir, &opts)
}

func recursiveDownload(
	ctx context.Context,
	client *terabox.Client,
	remotePaths []string,
	destDir string,
	opts *terabox.DownloadOptions,
) error {
	files, dirs := runner.List(ctx, client, remotePaths, terabox.OrderByName, false)

	if len(files.GetItems()) == 0 && len(dirs) == 0 {
		return fmt.Errorf("%s: No valid file or directory specified", getCmdName)
	}

	for _, file := range files.GetItems() {
		err := client.Download(file.Path, destDir, opts)
		if err != nil {
			return err
		}
	}

	for _, dir := range dirs {
		baseDir := path.Join(destDir, path.Base(path.Clean(dir.OriginalArg)))
		if stat, err := os.Stat(baseDir); err != nil {
			if err := os.Mkdir(baseDir, os.ModePerm); err != nil {
				return fmt.Errorf("%s: Failed to create directory", getCmdName)
			}
		} else if !stat.IsDir() {
			return fmt.Errorf("%s: \"%s\" is not directory", getCmdName, dir.OriginalArg)
		}

		var innerDirs []string
		for _, item := range dir.GetItems() {
			if item.IsDir == 0 {
				err := client.Download(item.Path, baseDir, opts)
				if err != nil {
					return err
				}
			} else {
				innerDirs = append(innerDirs, item.Path)
			}
		}

		if len(innerDirs) > 0 {
			if err := recursiveDownload(ctx, client, innerDirs, baseDir, opts); err != nil {
				return err
			}
		}
	}
	return nil
}
