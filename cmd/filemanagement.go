package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"tbc/internal/runner"
	"tbc/internal/terabox"
	"tbc/internal/util"

	"github.com/urfave/cli/v3"
)

const (
	moveCmdName   = "mv"
	copyCmdName   = "cp"
	removeCmdName = "rm"
	mkdirCmdName  = "mkdir"
	findCmdName   = "find"
)

var flagsForMoveCopy = []cli.Flag{
	&cli.BoolFlag{
		Name:        "force",
		Aliases:     []string{"f"},
		Usage:       "Force overwrite",
		HideDefault: true,
	},
	&cli.BoolFlag{
		Name:        "new-copy",
		Aliases:     []string{"n"},
		Usage:       "Make new copy file",
		HideDefault: true,
	},
}

var moveCmd = &cli.Command{
	Name:                   moveCmdName,
	Usage:                  "Move remote files or directories",
	UsageText:              fmt.Sprintf("%s %s SOURCE... DEST", CliName, moveCmdName),
	Action:                 moveAction,
	Flags:                  flagsForMoveCopy,
	UseShortOptionHandling: true,
}

var copyCmd = &cli.Command{
	Name:                   copyCmdName,
	Usage:                  "Copy remote files or directories",
	UsageText:              fmt.Sprintf("%s %s SOURCE... DEST", CliName, copyCmdName),
	Action:                 copyAction,
	Flags:                  flagsForMoveCopy,
	UseShortOptionHandling: true,
}

var removeCmd = &cli.Command{
	Name:      removeCmdName,
	Usage:     "Remove remote files or directories",
	UsageText: fmt.Sprintf("%s %s FILE...", CliName, removeCmdName),
	Action:    removeAction,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "interactive",
			Aliases: []string{"i"},
			Usage:   "Confirmation prompt before removing files",
			Value:   true,
		},
		&cli.BoolFlag{
			Name:        "force",
			Aliases:     []string{"f"},
			Usage:       "Force removal without confirmation prompt",
			HideDefault: true,
		},
	},
	UseShortOptionHandling: true,
}

var mkdirCmd = &cli.Command{
	Name:                   mkdirCmdName,
	Usage:                  "Make remote directory",
	UsageText:              fmt.Sprintf("%s %s DIRECTORY...", CliName, mkdirCmdName),
	Action:                 mkdirAction,
	Flags:                  []cli.Flag{},
	UseShortOptionHandling: true,
}

var findCmd = &cli.Command{
	Name:      findCmdName,
	Usage:     "Search for files in a remote directory",
	UsageText: fmt.Sprintf("%s %s PATTERN [DIRECTORY]", CliName, findCmdName),
	Action:    findAction,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:        "no-recursion",
			Aliases:     []string{"R"},
			Usage:       "Disable recursive search",
			Category:    "SEARCH OPTIONS",
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
	},
	UseShortOptionHandling: true,
}

func moveAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() < 2 {
		cli.ShowSubcommandHelp(cmd)
		return fmt.Errorf("%s: Invalid arguments", moveCmdName)
	}

	client, err := setupClient(cmd)
	if err != nil {
		return err
	}

	sources := cmd.Args().Slice()
	dest := sources[len(sources)-1]
	sources = sources[:len(sources)-1]

	sources = getTrueSources(ctx, client, sources)
	ondup := terabox.OpNone
	if cmd.Bool("new-copy") {
		ondup = terabox.OpNewCopy
	} else if cmd.Bool("force") {
		ondup = terabox.OpOverwrite
	}

	result, err := client.Move(sources, dest, ondup)
	if err != nil {
		return err
	}

	if !result {
		return fmt.Errorf("%s: Move failed", moveCmdName)
	}
	return nil
}

func copyAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() < 2 {
		cli.ShowSubcommandHelp(cmd)
		return fmt.Errorf("%s: Invalid arguments", copyCmdName)
	}

	client, err := setupClient(cmd)
	if err != nil {
		return err
	}

	sources := cmd.Args().Slice()
	dest := sources[len(sources)-1]
	sources = sources[:len(sources)-1]

	sources = getTrueSources(ctx, client, sources)

	ondup := terabox.OpNone
	if cmd.Bool("new-copy") {
		ondup = terabox.OpNewCopy
	} else if cmd.Bool("force") {
		ondup = terabox.OpOverwrite
	}

	result, err := client.Copy(sources, dest, ondup)
	if err != nil {
		return err
	}

	if !result {
		return fmt.Errorf("%s: Copy failed", copyCmdName)
	}
	return nil
}

func removeAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() < 1 {
		cli.ShowSubcommandHelp(cmd)
		return fmt.Errorf("%s: Invalid arguments", removeCmdName)
	}

	client, err := setupClient(cmd)
	if err != nil {
		return err
	}

	sources := cmd.Args().Slice()
	sources = getTrueSources(ctx, client, sources)
	if len(sources) < 1 {
		return fmt.Errorf("%s: No file or directory specified", removeCmdName)
	}

	if !cmd.Bool("force") {
		for _, source := range sources {
			fmt.Printf("remove %s? [Y/n]: ", source)
			reader := bufio.NewReader(os.Stdin)
			char, _, err := reader.ReadRune()
			if err != nil {
				continue
			}
			switch strings.ToLower(string(char)) {
			case "n":
				return fmt.Errorf("%s: User cancelled", removeCmdName)
			}
		}
	}

	result, err := client.Remove(sources)
	if err != nil {
		return err
	}

	if !result {
		return fmt.Errorf("%s: Remove failed", removeCmdName)
	}
	return nil
}

func mkdirAction(_ context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() < 1 {
		cli.ShowSubcommandHelp(cmd)
		return fmt.Errorf("%s: Invalid arguments", mkdirCmdName)
	}

	client, err := setupClient(cmd)
	if err != nil {
		return err
	}

	result := true
	for _, arg := range cmd.Args().Slice() {
		ok, err := client.MakeDir(arg)
		if err != nil {
			return err
		}
		result = result && ok
	}
	if !result {
		return fmt.Errorf("%s: MakeDir failed", mkdirCmdName)
	}
	return nil
}

func findAction(_ context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() < 1 {
		cli.ShowSubcommandHelp(cmd)
		return fmt.Errorf("%s: Invalid arguments", findCmdName)
	}

	client, err := setupClient(cmd)
	if err != nil {
		return err
	}

	pattern := cmd.Args().First()
	baseDir := "/"
	if cmd.Args().Len() > 1 {
		baseDir = cmd.Args().Slice()[1]
	}
	sort := terabox.OrderByName
	if cmd.Bool("size") {
		sort = terabox.OrderBySize
	} else if cmd.Bool("time") {
		sort = terabox.OrderByTime
	}
	rev := cmd.Bool("reverse")

	items, err := client.Search(terabox.SearchOptions{
		Keyword:     pattern,
		BaseDir:     baseDir,
		NoRecursion: cmd.Bool("no-recursion"),
		OrderBy:     sort,
		Desc:        rev,
	})
	if err != nil {
		return err
	}

	for _, item := range items {
		fmt.Println(item.Path)
	}
	return nil
}

func setupClient(cmd *cli.Command) (*terabox.Client, error) {
	cookie, err := util.GetCookie(cmd.Root().String(CookieFileOptName))
	if err != nil {
		return nil, err
	}

	// Create TeraBox client
	client, err := terabox.NewClient(cookie)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func getTrueSources(ctx context.Context, client *terabox.Client, sources []string) []string {
	files, dirs := runner.List(ctx, client, sources, terabox.OrderByName, false)
	trueSources := getFilePaths(files)
	for _, dir := range dirs {
		trueSources = append(trueSources, dir.Path)
	}

	return trueSources
}

func getFilePaths(items runner.ListResults) []string {
	result := []string{}
	for _, item := range items.GetItems() {
		result = append(result, item.Path)
	}
	return result
}
