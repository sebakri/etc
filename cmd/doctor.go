// Package cmd implements the command-line interface for box.
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/sebakri/box/internal/doctor"
)

// doctorCmd represents the doctor command
var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Checks if the host runtimes are installed",
	Run: func(_ *cobra.Command, _ []string) {
		doctor.Run()
	},
}

func init() {
	RootCmd.AddCommand(doctorCmd)
}
