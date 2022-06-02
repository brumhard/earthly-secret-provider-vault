package main

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

func buildVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: fmt.Sprintf("Print the version number of %s", cli),
		Long:  fmt.Sprintf("All software has versions. This is %s's.", cli),
		Run: func(cmd *cobra.Command, args []string) {
			if bi, ok := debug.ReadBuildInfo(); ok {
				fmt.Printf("%+v\n", bi)
			}
		},
	}

	return cmd
}
