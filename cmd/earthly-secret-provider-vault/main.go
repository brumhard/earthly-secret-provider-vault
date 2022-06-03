package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/brumhard/earthly-secret-provider-vault/pkg/provider"

	"github.com/moby/buildkit/session/secrets"
)

const cli = "earthly-secret-provider-vault"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cmd := buildRootCommand()
	if err := cmd.ExecuteContext(ctx); err != nil {
		if errors.Is(err, secrets.ErrNotFound) {
			// expected case, exit 2 indicates that next secret provider can be queried
			os.Exit(2)
		}

		fmt.Fprintf(os.Stderr, "An error occurred: %s\n", err)
		if errors.Is(err, provider.ErrInvalidConfig) {
			fmt.Fprintf(os.Stderr, "Please use %s config first to set required options or edit the config at %s directly.\n", cli, provider.CfgFilePath)
		}
		os.Exit(1)
	}
}
