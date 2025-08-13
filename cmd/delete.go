package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/vourteen14/ngcli/filesystem"
	"github.com/vourteen14/ngcli/utils"
)

var deleteForce bool

var deleteCmd = &cobra.Command{
	Use:   "delete <config_name>",
	Short: "Delete nginx configuration file",
	Long: `Delete nginx configuration file and remove any associated symlink.

The config name should be without the .conf extension.
Use --force to skip confirmation prompt.`,
	Args: cobra.ExactArgs(1),
	RunE: runDelete,
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	
	deleteCmd.Flags().BoolVar(&deleteForce, "force", false, "force deletion without confirmation")
}

func runDelete(cmd *cobra.Command, args []string) error {
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
	
	if !deleteForce {
		fmt.Printf("Are you sure you want to delete %s? (y/N): ", configName)
		var response string
		fmt.Scanln(&response)
		
		if response != "y" && response != "Y" && response != "yes" {
			fmt.Println("Deletion cancelled")
			return nil
		}
	}
	
	if enabledDir, hasEnabled := utils.DetectNginxEnabledPath(); hasEnabled {
		symlinkPath := filepath.Join(enabledDir, configName)
		if utils.FileExists(symlinkPath) {
			if err := filesystem.RemoveSymlink(symlinkPath); err != nil {
				fmt.Printf("Warning: failed to remove symlink %s: %v\n", symlinkPath, err)
			} else if verbose {
				fmt.Printf("Removed symlink: %s\n", symlinkPath)
			}
		}
	}
	
	if err := filesystem.DeleteFile(configPath); err != nil {
		return fmt.Errorf("failed to delete configuration: %w", err)
	}
	
	fmt.Printf("Deleted configuration: %s\n", configName)
	
	return nil
}