package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
)

func RunInteractive() error {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("tbc> ")

	// Customize CommandNotFound for interactive mode
	originalHandler := app.CommandNotFound
	app.CommandNotFound = func(ctx context.Context, c *cli.Command, s string) {
		fmt.Fprintf(os.Stderr, "Command not found: %s\n", s)
	}
	defer func() { app.CommandNotFound = originalHandler }()

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			fmt.Print("tbc> ")
			continue
		}

		if line == "exit" || line == "quit" {
			break
		}

		// Simple argument parsing (does not handle quotes yet)
		args := strings.Fields(line)

		// Prepend program name to args to satisfy urfave/cli requirements
		// app.Run expects [program_name, command, subcommands...]
		cmdArgs := append([]string{"tbc"}, args...)

		// Execute the command
		// We need to handle panics or exits if urfave/cli tries to exit
		// But urfave/cli v3 usually returns error instead of exiting unless configured otherwise
		err := app.Run(context.Background(), cmdArgs)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}

		fmt.Print("tbc> ")
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading standard input: %w", err)
	}

	return nil
}
