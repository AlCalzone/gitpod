package cmd

import (
	"fmt"

	config "github.com/gitpod-io/gitpod/installer/pkg/config/v1alpha1"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a base config file",
	Long: `Create a base config file

This file contains all the credentials to install a Gitpod instance and
be saved to a repository.`,
	Example: `  # Save config to config.yaml.
  gitpod-installer init > config.yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		var cfg config.Config
		fc, err := yaml.Marshal(cfg)
		if err != nil {
			panic(err)
		}

		fmt.Print(string(fc))
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
