package main

import (
	"os"

	"github.com/venkatkrishna07/mkdev/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
