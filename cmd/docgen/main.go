package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/brumhard/earthly-secret-provider-vault/cmd/earthly-secret-provider-vault/app"
	"github.com/spf13/cobra/doc"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "an error occurred: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	var output string
	flag.StringVar(&output, "o", "", "output directory")
	flag.Parse()

	if output == "" {
		return errors.New("output directory is required")
	}

	return doc.GenMarkdownTree(app.BuildRootCommand(), output)
}
