/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/bverschueren/in2un/pkg/reader"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// apiResourcesCmd represents the apiResources command
var apiResourcesCmd = &cobra.Command{
	Use:    "api-resources",
	Args:   cobra.MaximumNArgs(0),
	Short:  "(Experimental) List available resources in an Insights archive.",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		ir, err := reader.NewInsightsReader(viper.GetString("active"))
		if err != nil {
			log.Fatal(err)
		}
		found := ir.ReadResourceTypes()
		fmt.Printf("NAME\n")
		for f := range *found {
			fmt.Printf("%s\n", f)
		}
	},
}

func init() {
	InsightsCmd.AddCommand(apiResourcesCmd)
}
