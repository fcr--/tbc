package commands

import (
	"context"
	"fmt"
	"tbc/internal/util"

	"github.com/urfave/cli/v3"
)

const dfCmdName = "df"

var dfCmd = &cli.Command{
	Name:      dfCmdName,
	Usage:     "Display disk usage of TeraBox storage",
	UsageText: fmt.Sprintf("%s %s FILE", CliName, dfCmdName),
	Action:    dfAction,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:        "human-readable",
			Aliases:     []string{"H"},
			Usage:       "Print sized human readable form",
			HideDefault: true,
		},
		&cli.BoolFlag{
			Name:        "kibi-bytes",
			Aliases:     []string{"k"},
			Usage:       "Use 1024 bytes (1 Kibibyte)",
			HideDefault: true,
		},
	},
	UseShortOptionHandling: true,
}

func dfAction(ctx context.Context, cmd *cli.Command) error {
	// Create TeraBox client
	client, err := setupClient(cmd)
	if err != nil {
		return err
	}

	quota, err := client.GetQuota()
	if err != nil {
		return err
	}

	username, err := client.GetDisplayName()
	if err != nil {
		return err
	}

	sizes := []uint64{quota.Total, quota.Used, quota.Free, quota.SBoxUsed}
	names := []string{"Size", "Used", "Avail", "S-Box"}
	width := max(len(username), 7) + 2
	header := fmt.Sprintf("%-*s", width, "Account")
	value := fmt.Sprintf("%-*s", width, username)

	for i := range sizes {
		size := fmt.Sprintf("%d", sizes[i])
		if cmd.Bool("human-readable") {
			if cmd.Bool("kibi-bytes") {
				size = util.ByteCountIEC(int64(sizes[i])) + "iB"
			} else {
				size = util.ByteCountSI(int64(sizes[i])) + "B"
			}
		}
		name := names[i]

		maxLen := max(len(size), len(name)) + 2
		header += fmt.Sprintf("%*s", maxLen, name)
		value += fmt.Sprintf("%*s", maxLen, size)
	}

	fmt.Println(header)
	fmt.Println(value)

	return nil
}
