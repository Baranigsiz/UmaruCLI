package cmd

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// These variables are injected at build time via -ldflags by GoReleaser
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

var versionJSONFlag bool

type VersionInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

var versionCmd = &cobra.Command{
	Use:     "version",
	Short:   "Print the version number of Umaru",
	GroupID: "util",
	RunE: func(cmd *cobra.Command, args []string) error {
		if versionJSONFlag {
			info := VersionInfo{
				Version:   Version,
				Commit:    Commit,
				BuildDate: BuildDate,
				OS:        runtime.GOOS,
				Arch:      runtime.GOARCH,
			}
			data, err := json.MarshalIndent(info, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}
		fmt.Printf("Umaru CLI %s (commit: %s, built at: %s)\n", Version, Commit, BuildDate)
		return nil
	},
}

func init() {
	versionCmd.Flags().BoolVar(&versionJSONFlag, "json", false, "Output version info in JSON format")
	rootCmd.AddCommand(versionCmd)
}
