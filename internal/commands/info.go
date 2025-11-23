package commands

import (
	"context"
	"fmt"
	"github.com/urfave/cli/v3"
)

const infoCmdName = "info"

var infoCmd = &cli.Command{
	Name:                   infoCmdName,
	Usage:                  "Show user information",
	UsageText:              fmt.Sprintf("%s %s", CliName, infoCmdName),
	Action:                 infoAction,
	Flags:                  []cli.Flag{},
	UseShortOptionHandling: true,
}

func infoAction(ctx context.Context, cmd *cli.Command) error {
	// Create TeraBox client
	client, err := setupClient(cmd)
	if err != nil {
		return err
	}

	user, err := client.GetDisplayName()
	if err != nil {
		return err
	}

	isVip, err := client.IsVip()
	if err != nil {
		return err
	}

	fmt.Printf("Name:  %s\n", user)
	fmt.Printf("IsVip: %t\n", isVip)

	return nil
}
