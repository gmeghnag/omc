package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"

	"github.com/gmeghnag/omc/cmd/helpers"
	"github.com/gmeghnag/omc/types"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// contextsCmd represents the mg command
var MustGather = &cobra.Command{
	Use:     "mg",
	Aliases: []string{"must-gather", "must-gathers"},
	Short:   "List or delete previously used must-gathers.",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(0)
	},
}

var GetMustGather = &cobra.Command{
	Use:     "get",
	Aliases: []string{"ls", "list"},
	Short:   "List must-gathers saved in omc config file.",
	Run: func(cmd *cobra.Command, args []string) {
		file, _ := ioutil.ReadFile(viper.ConfigFileUsed())
		omcConfigJson := types.Config{}
		_ = json.Unmarshal([]byte(file), &omcConfigJson)

		var data [][]string
		var emptyData [][]string
		headers := []string{"current", "id", "path", "namespace"}
		var mg []types.Context
		mg = omcConfigJson.Contexts
		for _, context := range mg {
			_list := []string{context.Current, context.Id, context.Path, context.Project}
			data = append(data, _list)
		}
		if reflect.DeepEqual(data, emptyData) {
			fmt.Fprintln(os.Stderr, "There are no must-gather resources defined.")
			os.Exit(1)
		} else {
			helpers.PrintTable(headers, data)
		}
	},
}

func init() {

	MustGather.AddCommand(
		GetMustGather,
		DeleteCmd,
	)
}
