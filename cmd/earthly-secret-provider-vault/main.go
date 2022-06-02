package main

import (
	"earthly-vault-provider/pkg/provider"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const cli = "earthly-secret-provider-vault"

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
		Use:   cli,
		Short: fmt.Sprintf("%s is a secret provider for Earthly that connects to Vault", cli),
		Long: fmt.Sprintf(`%[1]s is a secret provider for Earthly that connects to Hashicorp's Vault.
For docs on how to configure this take a look here: https://docs.earthly.dev/docs/earthly-config#secret_provider-experimental.

Since the contract for secret providers is fairly simple you can test this provider by running:
		$ %[1]s <vault-path>
This print the secret on stdout.

Generally the CLI will look at ~/.vault-token and ~/.earthly/vault.yml for the configuration.
The token from ~/.vault-token will be used if it exists, otherwise the token from ~/.earthly/vault.yml will be used.
vault.yml should be used to set the Vault address and optionally a lookup secret can be added.

To set a config option in the vault.yml file, use the config subcommand.`, cli),
		// don't show errors and usage on errors in any RunE function.
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.ExactArgs(1),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Enable swapping out stdout/stderr for testing
			p.Out = cmd.OutOrStdout()
			p.Err = cmd.OutOrStderr()
		},
		RunE: func(_ *cobra.Command, args []string) error {
			return p.PrintSecret(args)
		},
	}

	cmd.AddCommand(buildVersionCommand())

	return cmd
}
