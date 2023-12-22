package alert

import (
	"fmt"
	"os"

	"github.com/gmeghnag/omc/cmd/prometheus"
	"github.com/spf13/cobra"
)

// alertCmd represents the alert command
var AlertCmd = &cobra.Command{
	Use:     "alert",
	Aliases: []string{"alerts"},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(os.Stderr, fmt.Errorf("Command \"omc alert\" is deprecated and will be removed in the next releases, use \"omc prometheus\" instead.\n"))
		prometheus.PrometheusCmd.Help()
	},
}
