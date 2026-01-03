package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vourteen14/ngcli/filesystem"
	"github.com/vourteen14/ngcli/system"
	"github.com/vourteen14/ngcli/template"
	"github.com/vourteen14/ngcli/utils"
)

var (
	setFlags     []string
	dryRun       bool
	output       string
	templateName string
	interactive  bool
)

var generateCmd = &cobra.Command{
	Use:   "generate <config_name>",
	Short: "Generate nginx configuration from template",
	Long: `Generate nginx configuration file from a template with specified parameters.

The config_name will be used as the output filename (config_name.conf).
After generation, the configuration will be automatically validated (nginx -t),
enabled (symlink created), and nginx will be reloaded.

If a file already exists, you will be prompted to confirm overwrite.
Use --dry-run to preview the configuration without writing files.

Examples:
  ngcli generate mysite --template prod --set domain=example.com
  ngcli generate api-server --template custom-api --set domain=api.example.com
  ngcli generate blog                    # Shows available templates to choose from
  ngcli generate test --dry-run          # Preview configuration without writing`,
	Args: cobra.ExactArgs(1),
	RunE: runGenerate,
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringArrayVar(&setFlags, "set", []string{}, "set template parameters (key=value)")
	generateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview output without writing files")
	generateCmd.Flags().StringVarP(&output, "output", "o", "", "override output file path")
	generateCmd.Flags().StringVarP(&templateName, "template", "t", "", "template to use (if not specified, shows available templates)")
	generateCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "interactive mode for parameter input")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	configName := args[0] 

	if templateName == "" {
		selectedTemplate, err := selectTemplate()
		if err != nil {
			return err
		}
		templateName = selectedTemplate
	}

	params, err := utils.ParseSetFlags(setFlags)
	if err != nil {
		return fmt.Errorf("failed to parse set flags: %w", err)
	}

	tmpl, err := template.LoadTemplate(templateName, templateDir)
	if err != nil {
		return fmt.Errorf("failed to load template: %w", err)
	}

	if interactive {
		params, err = interactiveParameterInput(tmpl, params)
		if err != nil {
			return fmt.Errorf("interactive input failed: %w", err)
		}
	}

	var content string
	if tmpl.Metadata != nil && len(tmpl.Metadata.Parameters) > 0 {
		if len(params) == 0 && !interactive && !dryRun {
			fmt.Printf("Template: %s\n", templateName)
			if tmpl.Metadata.Description != "" {
				fmt.Printf("Description: %s\n", tmpl.Metadata.Description)
			}
			fmt.Printf("\n%s", tmpl.Metadata.GetParameterHelp())
			
			fmt.Print("Enter parameters interactively? (Y/n): ")
			var response string
			if _, err := fmt.Scanln(&response); err != nil {
				// Treat scan error or empty input as "yes" (default)
				response = ""
			}

			if response == "" || response == "y" || response == "Y" || response == "yes" {
				interactive = true
			} else {
				fmt.Println("Use --set key=value to provide parameters manually")
				return nil
			}
		}

		if interactive {
			params, err = interactiveParameterInput(tmpl, params)
			if err != nil {
				return fmt.Errorf("interactive input failed: %w", err)
			}
		}

		if len(params) == 0 && dryRun {
			fmt.Printf("Template: %s\n", templateName)
			if tmpl.Metadata.Description != "" {
				fmt.Printf("Description: %s\n", tmpl.Metadata.Description)
			}
			fmt.Printf("\n%s", tmpl.Metadata.GetParameterHelp())
			fmt.Println("Use --set key=value to provide parameters, or --interactive for guided input")
			return nil
		}

		content, err = tmpl.RenderWithValidation(params)
		if err != nil {
			fmt.Printf("Template validation failed: %v\n\n", err)
			fmt.Printf("%s", tmpl.Metadata.GetParameterHelp())
			return fmt.Errorf("template validation failed")
		}
	} else {
		if err := validateRequiredParamsLegacy(templateName, params); err != nil {
			return err
		}
		content, err = tmpl.Render(params)
		if err != nil {
			return fmt.Errorf("failed to render template: %w", err)
		}
	}

	if dryRun {
		fmt.Printf("Config: %s (using template: %s)\n", configName, templateName)
		if tmpl.Metadata != nil && tmpl.Metadata.Description != "" {
			fmt.Printf("Description: %s\n", tmpl.Metadata.Description)
		}
		fmt.Println("Generated configuration preview:")
		fmt.Println(strings.Repeat("-", 50))
		fmt.Println(content)
		fmt.Println(strings.Repeat("-", 50))
		return nil
	}

	outputPath, err := getOutputPath(configName)
	if err != nil {
		return fmt.Errorf("failed to determine output path: %w", err)
	}

	// Check if file exists and prompt for overwrite
	if utils.FileExists(outputPath) {
		fmt.Printf("Configuration file already exists: %s\n", outputPath)
		fmt.Print("Overwrite existing file? (y/N): ")
		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			// Treat scan error or empty input as "no"
			fmt.Println("Operation cancelled")
			return nil
		}

		if response != "y" && response != "Y" && response != "yes" {
			fmt.Println("Operation cancelled")
			return nil
		}

		// Create backup before overwriting
		if err := filesystem.BackupFile(outputPath); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
		if verbose {
			fmt.Printf("Created backup of existing configuration\n")
		}
	}

	// Write configuration file
	if err := filesystem.WriteFile(outputPath, content, true); err != nil {
		return fmt.Errorf("failed to write configuration: %w", err)
	}

	fmt.Printf("Generated configuration: %s\n", outputPath)
	if tmpl.Metadata != nil && tmpl.Metadata.Description != "" {
		fmt.Printf("Template: %s - %s\n", templateName, tmpl.Metadata.Description)
	}

	// Run nginx -t validation
	if verbose {
		fmt.Println("Running nginx -t validation...")
	}

	if err := system.NginxTest(); err != nil {
		fmt.Printf("\nError: nginx -t validation failed: %v\n", err)
		fmt.Println("Configuration file generated but NOT enabled (syntax errors detected)")
		fmt.Println("Please fix the configuration manually and run 'ngcli enable' when ready")
		return fmt.Errorf("nginx validation failed")
	}

	if verbose {
		fmt.Println("nginx -t validation passed")
	}

	// Auto-enable: Create symlink in sites-enabled (Ubuntu/Debian only)
	enabledDir, hasEnabled := utils.DetectNginxEnabledPath()
	if hasEnabled {
		configFilename := filepath.Base(outputPath)
		sourcePath := outputPath
		targetPath := filepath.Join(enabledDir, configFilename)

		if err := filesystem.CreateSymlink(sourcePath, targetPath); err != nil {
			fmt.Printf("Warning: failed to enable configuration: %v\n", err)
			fmt.Println("Configuration generated but not enabled")
			fmt.Printf("Run 'ngcli enable %s' manually to enable it\n", configFilename)
			return nil
		}

		fmt.Printf("Enabled configuration: %s\n", configFilename)
		if verbose {
			fmt.Printf("Created symlink: %s -> %s\n", targetPath, sourcePath)
		}
	} else {
		if verbose {
			fmt.Println("sites-enabled directory not found (configuration is active by default on this system)")
		}
	}

	// Auto-reload nginx
	if verbose {
		fmt.Println("Reloading nginx configuration...")
	}

	if err := system.NginxReload(); err != nil {
		fmt.Printf("Warning: failed to reload nginx: %v\n", err)
		fmt.Println("Configuration generated and enabled but nginx reload failed")
		fmt.Println("Run 'ngcli reload' manually to apply changes")
		return nil
	}

	fmt.Println("Nginx configuration reloaded successfully")
	fmt.Println("\nConfiguration is now active!")

	return nil
}

func selectTemplate() (string, error) {
	templates, err := template.ListTemplates(templateDir)
	if err != nil {
		return "", fmt.Errorf("failed to list templates: %w", err)
	}

	if len(templates) == 0 {
		return "", fmt.Errorf("no templates found. Run 'ngcli init' to create default templates")
	}

	fmt.Println("Available templates:")
	for i, tmpl := range templates {
		if tmplObj, err := template.LoadTemplate(tmpl, templateDir); err == nil {
			description := tmplObj.Metadata.Description
			if description == "" {
				description = "No description"
			}
			fmt.Printf("  %d. %s - %s\n", i+1, tmpl, description)
		} else {
			fmt.Printf("  %d. %s\n", i+1, tmpl)
		}
	}

	fmt.Printf("\nSelect template (1-%d): ", len(templates))
	var choice int
	if _, err := fmt.Scanln(&choice); err != nil {
		return "", fmt.Errorf("invalid input")
	}

	if choice < 1 || choice > len(templates) {
		return "", fmt.Errorf("invalid choice: %d", choice)
	}

	return templates[choice-1], nil
}

func interactiveParameterInput(tmpl *template.Template, existingParams map[string]string) (map[string]string, error) {
	if tmpl.Metadata == nil || len(tmpl.Metadata.Parameters) == 0 {
		return existingParams, nil
	}

	params := make(map[string]string)
	
	for k, v := range existingParams {
		params[k] = v
	}

	fmt.Printf("\nInteractive parameter input for template: %s\n", tmpl.Name)
	if tmpl.Metadata.Description != "" {
		fmt.Printf("Description: %s\n", tmpl.Metadata.Description)
	}
	fmt.Println()

	for _, param := range tmpl.Metadata.Parameters {
		if _, exists := params[param.Name]; exists {
			continue
		}

		prompt := fmt.Sprintf("%s (%s)", param.Name, param.Description)
		if param.Default != "" {
			prompt += fmt.Sprintf(" [default: %s]", param.Default)
		}
		if param.Required {
			prompt += " *required*"
		}
		prompt += ": "

		fmt.Print(prompt)

		var value string
		if _, err := fmt.Scanln(&value); err != nil {
			// Treat scan error or empty input as empty string
			value = ""
		}

		if value == "" && param.Default != "" {
			value = param.Default
		}
		
		if value == "" && param.Required {
			fmt.Printf("Error: %s is required\n", param.Name)
			return nil, fmt.Errorf("missing required parameter: %s", param.Name)
		}
		
		if value != "" {
			params[param.Name] = value
		}
	}

	return params, nil
}

func validateRequiredParamsLegacy(templateName string, params map[string]string) error {
	var required []string

	switch templateName {
	case "prod":
		required = []string{"domain", "ssl_cert", "ssl_key", "root_path"}
	case "staging":
		required = []string{"domain", "root_path", "auth_file"}
	case "dev":
		required = []string{"domain", "root_path"}
	default:
		return nil
	}

	return utils.ValidateRequiredParams(params, required)
}

func getOutputPath(templateName string) (string, error) {
	if output != "" {
		return output, nil
	}

	var baseDir string
	if outputDir != "" {
		baseDir = outputDir
	} else {
		var err error
		baseDir, err = utils.DetectNginxConfigPath()
		if err != nil {
			return "", err
		}
	}

	filename := templateName + ".conf"
	return filepath.Join(baseDir, filename), nil
}