package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version string
	commit  string
	date    string
)

type Version struct {
	Version string
	Date    string
	Commit  string
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version and build information for karpenter-generate",
	Run: func(_ *cobra.Command, _ []string) {
		kgVersion := Version{
			Version: version,
			Date:    date,
			Commit:  commit,
		}
		verBytes, _ := json.Marshal(kgVersion)
		fmt.Println(string(verBytes))
	},
}

func init() {
	AddCommand(versionCmd)
}
