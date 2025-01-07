package insights

import (
	in2un "github.com/bverschueren/in2un/cmd"
)

var InsightsCmd = in2un.InsightsCmd

func init() {
	InsightsCmd.Use = "insights"
	in2un.ConfigDir = "$HOME/.omc/"
}
