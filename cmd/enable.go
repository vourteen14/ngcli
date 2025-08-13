package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/vourteen14/ngcli/filesystem"
	"github.com/vourteen14/ngcli/system"
	"github.com/vourteen14/ngcli/utils"
)

var enableNoReload bool

var enableCmd = &cobra.Command{
	Use:   "enable <config_name>",
	Short: "Enable nginx configuration by creating symlink",
	Long: `Enable nginx configuration by creating a symbolic link 
in sites-enabled directory and reload nginx configuration.

The config name should be without the .conf extension.
Use --no-reload to skip automatic nginx reload.`,
	Args: cobra.ExactArgs(1),
	RunE: runEnable,
}

func init() {
	rootCmd.AddCommand(enableCmd)
	
	enableCmd.Flags().BoolVar(&enableNoReload, "no-reload", false, "skip automatic nginx reload")
}

func runEnable(cmd *cobra.Command, args []string) error {
	configName := args[0]
	
	enabledDir, hasEnabled := utils.DetectNginxEnabledPath()
	if !hasEnabled {
		return fmt.Errorf("sites-enabled directory not found (this system may not support enable/disable)")
	}
	
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
	
	sourcePath := filepath.Join(configDir, configName)
	targetPath := filepath.Join(enabledDir, configName)
	
	if !utils.FileExists(sourcePath) {
		return fmt.Errorf("configuration file not found: %s", sourcePath)
	}
	
	if err := filesystem.CreateSymlink(sourcePath, targetPath); err != nil {
		return fmt.Errorf("failed to enable configuration: %w", err)
	}
	
	fmt.Printf("Enabled configuration: %s\n", configName)
	if verbose {
		fmt.Printf("Created symlink: %s -> %s\n", targetPath, sourcePath)
	}
	
	if !enableNoReload {
		if verbose {
			fmt.Println("Reloading nginx configuration")
		}
		
		if err := system.NginxReload(); err != nil {
			fmt.Printf("Warning: failed to reload nginx: %v\n", err)
			fmt.Println("Configuration enabled but nginx reload failed")
			fmt.Println("Run 'ngcli reload' manually to apply changes")
		} else {
			fmt.Println("Nginx configuration reloaded successfully")
		}
	}
	
	return nil
}