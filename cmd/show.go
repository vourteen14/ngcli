package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/vourteen14/ngcli/filesystem"
	"github.com/vourteen14/ngcli/utils"
)

var showCmd = &cobra.Command{
	Use:   "show <config_name>",
	Short: "Show contents of nginx configuration file",
	Long: `Display the contents of a nginx configuration file.

The config name should be without the .conf extension.`,
	Args: cobra.ExactArgs(1),
	RunE: runShow,
}

func init() {
	rootCmd.AddCommand(showCmd)
}

func runShow(cmd *cobra.Command, args []string) error {
	configName := args[0]
	
	var configDir string
	if outputDir != "" {
		configDir = outputDir
	} else {
		var err error
		configDir, err = utils.DetectNginxConfigPath()
		if err != nil {
			return fmt.Errorf("failed to detect nginx config directory: %w", err)
		}
	}
	
	configPath := filepath.Join(configDir, configName)
	
	if !utils.FileExists(configPath) {
		return fmt.Errorf("configuration file not found: %s", configPath)
	}
	
	content, err := filesystem.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read configuration: %w", err)
	}
	
	fmt.Printf("Configuration: %s\n", configPath)
	fmt.Println("---")
	fmt.Print(content)
	
	return nil
}