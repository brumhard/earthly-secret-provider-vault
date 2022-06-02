package main

import (
	"earthly-vault-provider/pkg/provider"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	cmd := buildRootCommand()
	if err := cmd.Execute(); err != nil {
		if errors.Is(err, provider.ErrNotFound) {
			// expected case, exit 2 indicates that next secret provider can be queried
			os.Exit(2)
		}

		fmt.Fprintf(os.Stderr, "an error occurred: %s\n", err)
		os.Exit(1)
	}
}

func buildRootCommand() *cobra.Command {
	p := provider.New()

	cmd := &cobra.Command{
		Use:   "earthly-secret-provider-vault",
		Short: "earthly-secret-provider-vault is a secret provider for Earthly that connects to Vault",
		Long: `earthly-secret-provider-vault is a secret provider for Earthly that connects to Hashicorp's Vault.
For docs on how to configure this take a look here: https://docs.earthly.dev/docs/earthly-config#secret_provider-experimental.

Since the contract for secret providers is fairly simple you can test this provider by running:
		$ earthly-secret-provider-vault <vault-path>
This print the secret on stdout.

Generally the CLI will look at ~/.vault-token and ~/.earthly/vault.yml for the configuration.
The token from ~/.vault-token will be used if it exists, otherwise the token from ~/.earthly/vault.yml will be used.
vault.yml should be used to set the Vault address and optionally a lookup secret can be added.

To set a config option in the vault.yml file, use the config subcommand.`,
		// don't show errors and usage on errors in any RunE function.
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return p.PrintSecret()
		},
	}

	return cmd
}
