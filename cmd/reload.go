package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vourteen14/ngcli/system"
)

var reloadDryRun bool
var testConfig bool

var reloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Reload nginx configuration",
	Long: `Reload nginx configuration to apply changes.

Use --test to test configuration syntax before reloading.
Use --dry-run to preview the command without executing it.`,
	RunE: runReload,
}

func init() {
	rootCmd.AddCommand(reloadCmd)
	
	reloadCmd.Flags().BoolVar(&reloadDryRun, "dry-run", false, "preview command without execution")
	reloadCmd.Flags().BoolVarP(&testConfig, "test", "t", false, "test configuration syntax before reloading")
}

func runReload(cmd *cobra.Command, args []string) error {
	if reloadDryRun {
		fmt.Println("Commands that would be executed:")
		if testConfig {
			fmt.Println("  nginx -t")
		}
		fmt.Println("  nginx -s reload")
		return nil
	}
	
	if testConfig {
		if verbose {
			fmt.Println("Testing nginx configuration syntax")
		}
		
		if err := system.NginxTest(); err != nil {
			return fmt.Errorf("configuration test failed: %w", err)
		}
		
		if verbose {
			fmt.Println("Configuration syntax is valid")
		}
	}
	
	if verbose {
		fmt.Println("Reloading nginx configuration")
	}
	
	if err := system.NginxReload(); err != nil {
		return fmt.Errorf("failed to reload nginx: %w", err)
	}
	
	fmt.Println("Nginx configuration reloaded successfully")
	
	return nil
}