package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/vourteen14/ngcli/filesystem"
	"github.com/vourteen14/ngcli/system"
	"github.com/vourteen14/ngcli/utils"
)

var disableNoReload bool

var disableCmd = &cobra.Command{
	Use:   "disable <config_name>",
	Short: "Disable nginx configuration by removing symlink",
	Long: `Disable nginx configuration by removing the symbolic link 
from sites-enabled directory and reload nginx configuration.

The config name should be without the .conf extension.
Use --no-reload to skip automatic nginx reload.`,
	Args: cobra.ExactArgs(1),
	RunE: runDisable,
}

func init() {
	rootCmd.AddCommand(disableCmd)
	
	disableCmd.Flags().BoolVar(&disableNoReload, "no-reload", false, "skip automatic nginx reload")
}

func runDisable(cmd *cobra.Command, args []string) error {
	configName := args[0]

	enabledDir, hasEnabled := utils.DetectNginxEnabledPath()
	if !hasEnabled {
		return fmt.Errorf("sites-enabled directory not found (this system may not support enable/disable)")
	}

	// Try to resolve the config name to get the actual filename
	// First check in sites-enabled directly
	targetPath := filepath.Join(enabledDir, configName)
	if !utils.FileExists(targetPath) {
		// Try with .conf extension
		targetPath = filepath.Join(enabledDir, configName+".conf")
		if !utils.FileExists(targetPath) {
			return fmt.Errorf("configuration not enabled: %s", configName)
		}
	}

	configFilename := filepath.Base(targetPath)
	
	if err := filesystem.RemoveSymlink(targetPath); err != nil {
		return fmt.Errorf("failed to disable configuration: %w", err)
	}
	
	fmt.Printf("Disabled configuration: %s\n", configFilename)
	if verbose {
		fmt.Printf("Removed symlink: %s\n", targetPath)
	}
	
	if !disableNoReload {
		if verbose {
			fmt.Println("Reloading nginx configuration")
		}
		
		if err := system.NginxReload(); err != nil {
			fmt.Printf("Warning: failed to reload nginx: %v\n", err)
			fmt.Println("Configuration disabled but nginx reload failed")
			fmt.Println("Run 'ngcli reload' manually to apply changes")
		} else {
			fmt.Println("Nginx configuration reloaded successfully")
		}
	}
	
	return nil
}