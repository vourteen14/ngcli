package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	templateDir string
	outputDir   string
	verbose     bool
)

var rootCmd = &cobra.Command{
	Use:   "ngcli",
	Short: "CLI tool for managing Nginx configurations",
	Long: `ngcli is a CLI tool that helps you generate, manage, and deploy 
Nginx configuration files using templates.

It supports multiple environments and provides commands to enable, 
disable, and reload configurations.`,
	Version: "1.0.0",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&templateDir, "template-dir", getDefaultTemplateDir(), "directory containing templates")
	rootCmd.PersistentFlags().StringVar(&outputDir, "output-dir", "", "override output directory")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}

func initConfig() {
	if verbose {
		fmt.Printf("Template directory: %s\n", templateDir)
		if outputDir != "" {
			fmt.Printf("Output directory: %s\n", outputDir)
		}
	}
}

func getDefaultTemplateDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".ngcli/templates"
	}
	return filepath.Join(homeDir, ".ngcli", "templates")
}