package cmd

import (
	"fmt"
	"os"

	"github.com/chazzy/g2o/cmd/lab"
	"github.com/chazzy/g2o/internal/root"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "g2o",
	Short: "g2o is a CLI application",
	Long:  `g2o is a CLI application built with Cobra.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return root.Run()
	},
}

func init() {
	rootCmd.AddCommand(lab.Command)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
