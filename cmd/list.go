package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/vourteen14/ngcli/filesystem"
	"github.com/vourteen14/ngcli/utils"
)

var listTemplates bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List nginx configurations or templates",
	Long: `List nginx configuration files in the output directory or 
available templates in the template directory.

Use --templates flag to list available templates instead of configurations.`,
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
	
	listCmd.Flags().BoolVarP(&listTemplates, "templates", "t", false, "list available templates")
}

func runList(cmd *cobra.Command, args []string) error {
	if listTemplates {
		return listAvailableTemplates()
	}
	
	return listConfigurations()
}

func listAvailableTemplates() error {
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		fmt.Printf("Template directory does not exist: %s\n", templateDir)
		fmt.Println("Run 'ngcli init' to initialize templates")
		return nil
	}
	
	entries, err := os.ReadDir(templateDir)
	if err != nil {
		return fmt.Errorf("failed to read template directory: %w", err)
	}
	
	var templates []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		name := entry.Name()
		if filepath.Ext(name) == ".tpl" {
			templateName := name
			if len(name) > 9 && name[len(name)-9:] == ".conf.tpl" {
				templateName = name[:len(name)-9]
			}
			templates = append(templates, templateName)
		}
	}
	
	if len(templates) == 0 {
		fmt.Printf("No templates found in %s\n", templateDir)
		return nil
	}
	
	fmt.Printf("Available templates (%s):\n", templateDir)
	fmt.Printf("%-20s %s\n", "NAME", "FILE")
	fmt.Printf("%-20s %s\n", "----", "----")
	
	for _, tmpl := range templates {
		filename := tmpl + ".conf.tpl"
		fmt.Printf("%-20s %s\n", tmpl, filename)
	}
	
	fmt.Printf("\nTotal: %d templates\n", len(templates))
	
	return nil
}

func listConfigurations() error {
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
	
	configs, err := filesystem.ListConfigs(configDir)
	if err != nil {
		return fmt.Errorf("failed to list configurations: %w", err)
	}
	
	if len(configs) == 0 {
		fmt.Printf("No configuration files found in %s\n", configDir)
		return nil
	}
	
	var tableData [][]string
	enabledDir, hasEnabled := utils.DetectNginxEnabledPath()
	
	for _, config := range configs {
		// Strip .conf extension for display consistency
		displayName := config
		if len(config) > 5 && config[len(config)-5:] == ".conf" {
			displayName = config[:len(config)-5]
		}

		status := "disabled"
		if hasEnabled {
			enabledPath := filepath.Join(enabledDir, config)
			if utils.FileExists(enabledPath) {
				status = "enabled"
			}
		} else {
			status = "n/a"
		}
		tableData = append(tableData, []string{displayName, status})
	}
	
	fmt.Printf("Nginx configurations (%s):\n", configDir)
	fmt.Printf("%-30s %s\n", "NAME", "STATUS")
	fmt.Printf("%-30s %s\n", "----", "------")
	
	for _, row := range tableData {
		fmt.Printf("%-30s %s\n", row[0], row[1])
	}
	
	fmt.Printf("\nTotal: %d configurations\n", len(configs))
	
	return nil
}