package getsource

import (
	"os"

	"github.com/spf13/cobra"
)

var Image, AuthFile, FileName string

var GetSource = &cobra.Command{
	Use:     "source",
	Aliases: []string{"get-source"},
	Short:   "Retrieve OpenShift container image manifest and source code.",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(0)
	},
}

func init() {
	GetSource.AddCommand(
		GetCode,
		GetManifest,
	)
}
