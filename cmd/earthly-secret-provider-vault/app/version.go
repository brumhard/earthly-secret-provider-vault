package app

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

func BuildVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: fmt.Sprintf("Print the version number of %s", CLI),
		Long:  fmt.Sprintf("All software has versions. This is %s's.", CLI),
		Run: func(cmd *cobra.Command, args []string) {
			if bi, ok := debug.ReadBuildInfo(); ok {
				fmt.Printf("%+v\n", bi)
			}
		},
	}

	return cmd
}
