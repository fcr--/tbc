package main

import (
	"fmt"
	"os"
	"tbc/cmd"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
