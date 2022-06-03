package main

import (
	"earthly-vault-provider/pkg/provider"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

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

For configuration you can also use the config command:
		$ vault login --method=userpass username=test
		$ %[1]s config token $(vault print token)
		$ %[1]s config address $VAULT_ADDR

To set a config option in the vault.yml file, use the config subcommand.`, cli),
		// don't show errors and usage on errors in any RunE function.
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.ExactArgs(1),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Enable swapping out stdout/stderr for testing
			p.Logger = log.New(cmd.OutOrStderr(), "", 0)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			secretFetcher, err := p.LoadSecretStore()
			if err != nil {
				return err
			}

			secret, err := secretFetcher.GetSecret(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			fmt.Fprint(cmd.OutOrStdout(), string(secret))
			return nil
		},
	}

	cmd.AddCommand(buildVersionCommand())
	cmd.AddCommand(buildConfigCommand(p))

	return cmd
}