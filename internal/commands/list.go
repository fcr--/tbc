package commands

import (
	"context"
	"fmt"
	"os"
	"tbc/internal/formatter"
	"tbc/internal/runner"
	"tbc/internal/terabox"

	"github.com/urfave/cli/v3"
)

const listCmdName = "ls"

var listCmd = &cli.Command{
	Name:      listCmdName,
	Usage:     "List files in a remote directory",
	UsageText: fmt.Sprintf("%s %s [OPTIONS]... [FILE]...", CliName, listCmdName),
	Action:    listAction,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:        "oneline",
			Aliases:     []string{"1"},
			Usage:       "Single column output",
			Category:    "DISPLAY OPTIONS",
			HideDefault: true,
		},
		&cli.BoolFlag{
			Name:        "long",
			Aliases:     []string{"l"},
			Usage:       "Display extended file metadata as a table",
			Category:    "DISPLAY OPTIONS",
			HideDefault: true,
		},
		&cli.BoolFlag{
			Name:        "across",
			Aliases:     []string{"x"},
			Usage:       "Sort horizontally",
			Category:    "DISPLAY OPTIONS",
			HideDefault: true,
		},
		&cli.BoolFlag{
			Name:        "indicator",
			Aliases:     []string{"F"},
			Usage:       "Disable file type indicator",
			Category:    "DISPLAY OPTIONS",
			HideDefault: true,
		},
		&cli.BoolFlag{
			Name:        "only-dirs",
			Aliases:     []string{"D"},
			Usage:       "Print only directories",
			Category:    "FILTER AND SORTING OPTIONS",
			HideDefault: true,
		},
		&cli.BoolFlag{
			Name:        "only-files",
			Aliases:     []string{"f"},
			Usage:       "Print only files",
			Category:    "FILTER AND SORTING OPTIONS",
			HideDefault: true,
		},
		&cli.BoolFlag{
			Name:        "size",
			Aliases:     []string{"S"},
			Usage:       "Sort by size",
			Category:    "FILTER AND SORTING OPTIONS",
			HideDefault: true,
		},
		&cli.BoolFlag{
			Name:        "time",
			Aliases:     []string{"t"},
			Usage:       "Sort by modified time",
			Category:    "FILTER AND SORTING OPTIONS",
			HideDefault: true,
		},
		&cli.BoolFlag{
			Name:        "reverse",
			Aliases:     []string{"r"},
			Usage:       "Reverse sort order",
			Category:    "FILTER AND SORTING OPTIONS",
			HideDefault: true,
		},
		&cli.BoolFlag{
			Name:        "human-readable",
			Aliases:     []string{"H"},
			Usage:       "Print sized human readable form",
			Category:    "LONG VIEW OPTIONS",
			HideDefault: true,
		},
		&cli.BoolFlag{
			Name:        "id",
			Aliases:     []string{"i"},
			Usage:       "Print file ID",
			Category:    "LONG VIEW OPTIONS",
			HideDefault: true,
		},
		&cli.BoolFlag{
			Name:        "created",
			Aliases:     []string{"U"},
			Usage:       "File creation time",
			Category:    "LONG VIEW OPTIONS",
			HideDefault: true,
		},
		&cli.BoolFlag{
			Name:     "modified",
			Usage:    "File modified time",
			Category: "LONG VIEW OPTIONS",
			Value:    true,
		},
	},
	UseShortOptionHandling: true,
}

func listAction(ctx context.Context, cmd *cli.Command) error {
	formatOpts := formatter.OptsLongFormat{
		HumanReadable: cmd.Bool("human-readable"),
		Indicator:     !cmd.Bool("indicator"),
		CreationTime:  cmd.Bool("created"),
		FileId:        cmd.Bool("id"),
	}
	displayOpts := OptDisplayName{
		Indicator:    !cmd.Bool("indicator"),
		OnlyFiles:    cmd.Bool("only-files"),
		OnlyDirs:     cmd.Bool("only-dirs"),
		AbsolutePath: cmd.Bool("absolute"),
		Across:       cmd.Bool("across"),
		Margin:       2,
	}
	sort := terabox.OrderByName
	if cmd.Bool("size") {
		sort = terabox.OrderBySize
	} else if cmd.Bool("time") {
		sort = terabox.OrderByTime
	}
	rev := cmd.Bool("reverse")

	// Create TeraBox client
	client, err := setupClient(cmd)
	if err != nil {
		return err
	}

	// Get list
	files, dirs := runner.List(ctx, client, cmd.Args().Slice(), sort, rev)

	if cmd.Bool("long") {
		// Long list
		printLongStyle(files, &displayOpts, &formatOpts)
		for i, dir := range dirs {
			if len(dirs) > 1 && files.Len() == 0 {
				if i > 0 {
					fmt.Println()
				}
				fmt.Printf("%s:\n", dir.OriginalArg)
			}
			printLongStyle(dir, &displayOpts, &formatOpts)
		}
	} else {
		// Grid list
		if cmd.Bool("oneline") {
			displayOpts.Margin = 1000000
		}
		printGridStyle(files, &displayOpts)

		for i, dir := range dirs {
			if len(dirs) > 1 && files.Len() == 0 {
				if i > 0 {
					fmt.Println()
				}
				fmt.Printf("%s:\n", dir.OriginalArg)
			}
			printGridStyle(dir, &displayOpts)
		}
	}

	return nil
}

type OptDisplayName struct {
	Indicator    bool
	OnlyFiles    bool
	OnlyDirs     bool
	AbsolutePath bool
	Across       bool
	Margin       int
}

func printGridStyle(items runner.ListResults, opts *OptDisplayName) {
	displayNames := []string{}
	for i, item := range items.GetItems() {
		if item.IsDir == 0 && opts.OnlyDirs && !opts.OnlyFiles {
			continue
		} else if item.IsDir != 0 && !opts.OnlyDirs && opts.OnlyFiles {
			continue
		}
		displayName := items.GetDisplayPath(i)
		if opts.Indicator && item.IsDir != 0 {
			displayName += "/"
		}
		displayNames = append(displayNames, displayName)
	}
	formatter.FprintGrid(os.Stdout, displayNames, opts.Margin, opts.Across)
}

func printLongStyle(items runner.ListResults, dispOpts *OptDisplayName, fmtOpts *formatter.OptsLongFormat) {
	for _, item := range items.GetItems() {
		if item.IsDir == 0 && dispOpts.OnlyDirs && !dispOpts.OnlyFiles {
			continue
		} else if item.IsDir != 0 && !dispOpts.OnlyDirs && dispOpts.OnlyFiles {
			continue
		}
		formatter.FprintLong(os.Stdout, item, fmtOpts)
	}
}
