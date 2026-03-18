package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "rukkie",
	Short: "RukkiePulse — CLI observability for your backend services",
}

var envFlag string

func init() {
	rootCmd.PersistentFlags().StringVarP(&envFlag, "env", "e", "dev", "environment to use from rukkie.yaml")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
