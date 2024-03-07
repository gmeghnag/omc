package nodelogs

import (
	"fmt"
	"os"
	"strings"

	"github.com/gmeghnag/omc/vars"
	"github.com/spf13/cobra"
)

var NodeLogs = &cobra.Command{
	Use: "node-logs",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("The following node services logs are available to be read:")
			fmt.Println("")
			files, _ := os.ReadDir(vars.MustGatherRootPath + "/host_service_logs/masters/")
			for _, f := range files {
				fmt.Println("-", strings.TrimSuffix(f.Name(), "_service.log"))
			}
			fmt.Println("\nis it possible to read the content by executing 'omc node-logs <SERVICE>'.")
		}
		if len(args) > 1 {
			fmt.Fprintln(os.Stderr, "Expect zero arguemnt, found: ", len(args))
			os.Exit(1)
		}
		if len(args) == 1 {
			text, err := os.ReadFile(vars.MustGatherRootPath + "/host_service_logs/masters/" + args[0] + "_service.log")
			if err != nil {
				fmt.Fprintln(os.Stderr, "logs for service \""+args[0]+"\" not found or readable.")
				os.Exit(1)
			}
			fmt.Print(string(text))
		}
	},
}
