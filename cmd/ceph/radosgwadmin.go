package ceph

import (
	"os"

	"github.com/gmeghnag/omc/vars"
	"github.com/spf13/cobra"
)

var RadosgwAdmin = &cobra.Command{
	Use:                "radosgw-admin [command] [args...]",
	Short:              "Shows pre-captured radosgw-admin command output from an ODF must-gather.",
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
			cmd.Help()
			os.Exit(0)
		}
		LookupAndPrint(vars.MustGatherRootPath, "radosgw-admin", args)
	},
}
