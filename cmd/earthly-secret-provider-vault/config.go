package main

import (
	"fmt"

	"earthly-vault-provider/pkg/provider"

	"github.com/spf13/cobra"
)

func buildConfigCommand(p *provider.Provider) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: fmt.Sprintf("Set a configuration value for %s", cli),
		Long: fmt.Sprintf(`Set a configuration value for %[1]s.
Support config fields to be set are "address", "token" and "prefix".
This must be used to set the Vault address and token prior to using the secret provider,
since currently the secret providers cannot read any env vars.

To set the vault address in a vault aware system for example do:
	$ %[1]s config address $VAULT_ADDR
	
To unset config values you can just use the zero value, like for example
	$ %[1]s config address ""`, cli),
		Args:      cobra.ExactArgs(2),
		ValidArgs: []string{"address", "token", "prefix"},
		RunE: func(_ *cobra.Command, args []string) error {
			return p.SetConfigKey(args[0], args[1])
		},
	}

	return cmd
}
