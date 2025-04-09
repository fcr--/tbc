package cmd

import (
	"context"
	"os"

	"github.com/urfave/cli/v3"
)

const (
	CliName           = "tbc"
	CookieFileOptName = "cookie-file"
)

var app = &cli.Command{
	Name:    CliName,
	Authors: []any{"SHA-5010"},
	Usage:   "TeraBox CLI client",
	CommandNotFound: func(ctx context.Context, c *cli.Command, s string) {
		cli.ShowAppHelpAndExit(c, 3)
	},
	EnableShellCompletion: true,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    CookieFileOptName,
			Usage:   "TeraBox cookie file\n # if not specified, use TERABOX_COOKIE environment variable",
			Aliases: []string{"c"},
		},
	},
	Commands: []*cli.Command{
		listCmd,
		moveCmd,
		copyCmd,
		removeCmd,
		mkdirCmd,
		findCmd,
		infoCmd,
		putCmd,
		getCmd,
		dfCmd,
	},
}

func Execute() error {
	return app.Run(context.Background(), os.Args)
}
