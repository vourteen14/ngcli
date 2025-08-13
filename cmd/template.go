package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vourteen14/ngcli/filesystem"
	"github.com/vourteen14/ngcli/template"
	"github.com/vourteen14/ngcli/utils"
)

var (
	fromTemplate string
	editorFlag   string
	showParams   bool
)

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Manage nginx configuration templates",
	Long: `Manage nginx configuration templates including creating, editing,
listing, and validating templates.

Templates use comment-based metadata for parameter definitions.`,
}

var templateCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new template",
	Long: `Create a new nginx configuration template.

Use --from flag to create from existing template.`,
	Args: cobra.ExactArgs(1),
	RunE: runTemplateCreate,
}

var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available templates",
	Long:  `List all available templates including built-in and custom templates.`,
	RunE:  runTemplateList,
}

var templateShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show template content and metadata",
	Long: `Show template content and parameter information.

Use --params flag to show only parameter information.`,
	Args: cobra.ExactArgs(1),
	RunE: runTemplateShow,
}

var templateEditCmd = &cobra.Command{
	Use:   "edit <name>",
	Short: "Edit template in text editor",
	Long: `Edit template using your preferred text editor.

Editor selection priority:
  1. --editor flag
  2. $VISUAL environment variable
  3. $EDITOR environment variable
  4. System default (nano, vi, vim)`,
	Args: cobra.ExactArgs(1),
	RunE: runTemplateEdit,
}

var templateDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a custom template",
	Long: `Delete a custom template. Built-in templates (prod, staging, dev) 
cannot be deleted.`,
	Args: cobra.ExactArgs(1),
	RunE: runTemplateDelete,
}

var templateValidateCmd = &cobra.Command{
	Use:   "validate <name>",
	Short: "Validate template syntax",
	Long:  `Validate template syntax and metadata format.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplateValidate,
}

func init() {
	rootCmd.AddCommand(templateCmd)
	
	templateCmd.AddCommand(templateCreateCmd)
	templateCmd.AddCommand(templateListCmd)
	templateCmd.AddCommand(templateShowCmd)
	templateCmd.AddCommand(templateEditCmd)
	templateCmd.AddCommand(templateDeleteCmd)
	templateCmd.AddCommand(templateValidateCmd)
	
	templateCreateCmd.Flags().StringVar(&fromTemplate, "from", "", "create template from existing template")
	templateEditCmd.Flags().StringVar(&editorFlag, "editor", "", "text editor to use")
	templateShowCmd.Flags().BoolVar(&showParams, "params", false, "show only parameter information")
}

func runTemplateCreate(cmd *cobra.Command, args []string) error {
	templateName := args[0]
	templatePath := filepath.Join(templateDir, templateName+".conf.tpl")
	
	if utils.FileExists(templatePath) {
		return fmt.Errorf("template already exists: %s", templateName)
	}
	
	var content string
	
	if fromTemplate != "" {
		sourceTemplate, err := template.LoadTemplate(fromTemplate, templateDir)
		if err != nil {
			return fmt.Errorf("failed to load source template: %w", err)
		}
		
		content = updateTemplateMetadata(sourceTemplate.Content, templateName, fromTemplate)
	} else {
		content = generateBasicTemplate(templateName)
	}
	
	if err := filesystem.WriteFile(templatePath, content, false); err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}
	
	fmt.Printf("Created template: %s\n", templateName)
	fmt.Printf("Template file: %s\n", templatePath)
	
	fmt.Print("Open template in editor now? (Y/n): ")
	var response string
	fmt.Scanln(&response)
	
	if response == "" || response == "y" || response == "Y" || response == "yes" {
		editor := detectEditor(editorFlag)
		
		if verbose {
			fmt.Printf("Opening template with %s\n", editor)
		}
		
		editCmd := prepareEditorCommand(editor, templatePath)
		editCmd.Stdin = os.Stdin
		editCmd.Stdout = os.Stdout
		editCmd.Stderr = os.Stderr
		
		if err := editCmd.Run(); err != nil {
			fmt.Printf("Warning: editor failed: %v\n", err)
		} else {
			if err := template.ValidateTemplate(templatePath); err != nil {
				fmt.Printf("Warning: template validation failed: %v\n", err)
				fmt.Println("Template saved but may have syntax errors")
			} else {
				fmt.Printf("Template %s saved and validated successfully\n", templateName)
			}
		}
	} else {
		fmt.Println("Template created. Use 'ngcli template edit' to modify it later")
	}
	
	return nil
}

func runTemplateList(cmd *cobra.Command, args []string) error {
	templates, err := template.ListTemplates(templateDir)
	if err != nil {
		return fmt.Errorf("failed to list templates: %w", err)
	}
	
	if len(templates) == 0 {
		fmt.Printf("No templates found in %s\n", templateDir)
		fmt.Println("Run 'ngcli init' to create default templates or 'ngcli template create' to create custom templates")
		return nil
	}
	
	var tableData [][]string
	
	for _, tmplName := range templates {
		tmpl, err := template.LoadTemplate(tmplName, templateDir)
		if err != nil {
			tableData = append(tableData, []string{tmplName, "error", "failed to parse"})
			continue
		}
		
		description := tmpl.Metadata.Description
		if description == "" {
			description = "no description"
		}
		
		if len(description) > 50 {
			description = description[:47] + "..."
		}
		
		templateType := "custom"
		if tmplName == "prod" || tmplName == "staging" || tmplName == "dev" {
			templateType = "built-in"
		}
		
		tableData = append(tableData, []string{tmplName, templateType, description})
	}
	
	fmt.Printf("Available templates (%s):\n", templateDir)
	fmt.Printf("%-20s %-10s %s\n", "NAME", "TYPE", "DESCRIPTION")
	fmt.Printf("%-20s %-10s %s\n", "----", "----", "-----------")
	
	for _, row := range tableData {
		fmt.Printf("%-20s %-10s %s\n", row[0], row[1], row[2])
	}
	
	fmt.Printf("\nTotal: %d templates\n", len(templates))
	
	return nil
}

func runTemplateShow(cmd *cobra.Command, args []string) error {
	templateName := args[0]
	
	tmpl, err := template.LoadTemplate(templateName, templateDir)
	if err != nil {
		return fmt.Errorf("failed to load template: %w", err)
	}
	
	if showParams {
		fmt.Printf("Template: %s\n", templateName)
		if tmpl.Metadata.Description != "" {
			fmt.Printf("Description: %s\n", tmpl.Metadata.Description)
		}
		fmt.Printf("\n%s", tmpl.Metadata.GetParameterHelp())
	} else {
		fmt.Printf("Template: %s\n", templateName)
		fmt.Printf("File: %s\n", tmpl.Path)
		if tmpl.Metadata.Description != "" {
			fmt.Printf("Description: %s\n", tmpl.Metadata.Description)
		}
		if tmpl.Metadata.Author != "" {
			fmt.Printf("Author: %s\n", tmpl.Metadata.Author)
		}
		if tmpl.Metadata.Version != "" {
			fmt.Printf("Version: %s\n", tmpl.Metadata.Version)
		}
		
		fmt.Printf("\n%s", tmpl.Metadata.GetParameterHelp())
		
		fmt.Println("Template content:")
		fmt.Println(strings.Repeat("-", 50))
		fmt.Print(tmpl.Content)
		fmt.Println(strings.Repeat("-", 50))
	}
	
	return nil
}

func runTemplateEdit(cmd *cobra.Command, args []string) error {
	templateName := args[0]
	templatePath := filepath.Join(templateDir, templateName+".conf.tpl")
	
	if !utils.FileExists(templatePath) {
		return fmt.Errorf("template not found: %s", templateName)
	}
	
	editor := detectEditor(editorFlag)
	
	if verbose {
		fmt.Printf("Opening template %s with %s\n", templateName, editor)
	}
	
	editCmd := prepareEditorCommand(editor, templatePath)
	editCmd.Stdin = os.Stdin
	editCmd.Stdout = os.Stdout
	editCmd.Stderr = os.Stderr
	
	if err := editCmd.Run(); err != nil {
		return fmt.Errorf("editor failed: %w", err)
	}
	
	if err := template.ValidateTemplate(templatePath); err != nil {
		fmt.Printf("Warning: template validation failed: %v\n", err)
		fmt.Println("Template saved but may have syntax errors")
	} else {
		fmt.Printf("Template %s saved and validated successfully\n", templateName)
	}
	
	return nil
}

func runTemplateDelete(cmd *cobra.Command, args []string) error {
	templateName := args[0]
	
	builtInTemplates := []string{"prod", "staging", "dev"}
	for _, builtin := range builtInTemplates {
		if templateName == builtin {
			return fmt.Errorf("cannot delete built-in template: %s", templateName)
		}
	}
	
	templatePath := filepath.Join(templateDir, templateName+".conf.tpl")
	
	if !utils.FileExists(templatePath) {
		return fmt.Errorf("template not found: %s", templateName)
	}
	
	fmt.Printf("Are you sure you want to delete template '%s'? (y/N): ", templateName)
	var response string
	fmt.Scanln(&response)
	
	if response != "y" && response != "Y" && response != "yes" {
		fmt.Println("Deletion cancelled")
		return nil
	}
	
	if err := filesystem.DeleteFile(templatePath); err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}
	
	fmt.Printf("Deleted template: %s\n", templateName)
	
	return nil
}

func runTemplateValidate(cmd *cobra.Command, args []string) error {
	templateName := args[0]
	
	tmpl, err := template.LoadTemplate(templateName, templateDir)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	
	fmt.Printf("Template: %s\n", templateName)
	fmt.Printf("Syntax: valid\n")
	
	if len(tmpl.Metadata.Parameters) > 0 {
		fmt.Printf("Parameters: %d defined\n", len(tmpl.Metadata.Parameters))
		
		var required []string
		for _, param := range tmpl.Metadata.Parameters {
			if param.Required {
				required = append(required, param.Name)
			}
		}
		
		if len(required) > 0 {
			fmt.Printf("Required parameters: %s\n", strings.Join(required, ", "))
		}
	} else {
		fmt.Printf("Parameters: none defined\n")
	}
	
	fmt.Println("Template validation successful")
	
	return nil
}

func detectEditor(editorFlag string) string {
	if editorFlag != "" {
		return editorFlag
	}
	
	if visual := os.Getenv("VISUAL"); visual != "" {
		return visual
	}
	
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	
	editors := []string{"nano", "vi", "vim", "code", "emacs"}
	for _, editor := range editors {
		if _, err := exec.LookPath(editor); err == nil {
			return editor
		}
	}
	
	return "vi" 
}

func prepareEditorCommand(editor, filePath string) *exec.Cmd {
	switch {
	case strings.Contains(editor, "code"):
		if !strings.Contains(editor, "--wait") {
			editor += " --wait"
		}
	case strings.Contains(editor, "subl"):
		if !strings.Contains(editor, "--wait") {
			editor += " --wait"
		}
	}
	
	return exec.Command("sh", "-c", fmt.Sprintf("%s %s", editor, filePath))
}

func updateTemplateMetadata(content, newName, sourceName string) string {
	lines := strings.Split(content, "\n")
	var result []string
	
	for _, line := range lines {
		if strings.Contains(line, "# Template:") {
			result = append(result, fmt.Sprintf("# Template: %s", newName))
		} else if strings.Contains(line, "# Description:") {
			result = append(result, fmt.Sprintf("# Description: %s (based on %s)", newName, sourceName))
		} else {
			result = append(result, line)
		}
	}
	
	return strings.Join(result, "\n")
}

func generateBasicTemplate(templateName string) string {
	return fmt.Sprintf(`# Template: %s
# Description: Custom nginx configuration
# Author: %s
# Version: 1.0
#
# @param domain string required "Primary domain"
# @param port integer optional "Server port" default=80
# @param root_path string required "Document root path"

server {
    listen {{.port}};
    server_name {{.domain}};
    
    root {{.root_path}};
    index index.html index.htm;
    
    location / {
        try_files $uri $uri/ =404;
    }
}`, templateName, os.Getenv("USER"))
}