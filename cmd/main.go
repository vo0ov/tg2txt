package main

import (
	"os"

	"github.com/vo0ov/tg2txt/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}
