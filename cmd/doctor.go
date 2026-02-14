package cmd

import (
	"github.com/sebakri/box/internal/doctor"
	"github.com/spf13/cobra"
)

// doctorCmd represents the doctor command
var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Checks if the host runtimes are installed",
	Run: func(cmd *cobra.Command, args []string) {
		doctor.Run()
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
