package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var helpCmd = &cobra.Command{
	Use:   "help [command]",
	Short: "Display help information for ngcli commands",
	Long: `Display detailed help information and examples for ngcli commands.

Examples:
  ngcli help          Show general help
  ngcli help generate Show help for generate command
  ngcli help init     Show help for init command`,
	Run: runHelp,
}

func init() {
	rootCmd.AddCommand(helpCmd)
}

func runHelp(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		showGeneralHelp()
		return
	}

	commandName := args[0]
	switch commandName {
	case "init":
		showInitHelp()
	case "generate":
		showGenerateHelp()
	case "list":
		showListHelp()
	case "show":
		showShowHelp()
	case "enable":
		showEnableHelp()
	case "disable":
		showDisableHelp()
	case "delete":
		showDeleteHelp()
	case "reload":
		showReloadHelp()
	case "template":
		showTemplateHelp()
	default:
		fmt.Printf("Unknown command: %s\n", commandName)
		fmt.Println("Run 'ngcli help' to see available commands")
	}
}

func showGeneralHelp() {
	fmt.Println(`ngcli - Nginx Configuration CLI Tool

USAGE:
  ngcli [command] [flags]

AVAILABLE COMMANDS:
  init        Initialize ngcli with default templates
  generate    Generate nginx configuration from template
  list        List nginx configurations or templates
  show        Show contents of nginx configuration file
  enable      Enable nginx configuration (create symlink + reload)
  disable     Disable nginx configuration (remove symlink + reload)
  delete      Delete nginx configuration file
  reload      Reload nginx configuration
  template    Manage nginx configuration templates
  help        Display help information

GLOBAL FLAGS:
  --template-dir string   Directory containing templates
  --output-dir string     Override output directory
  -v, --verbose          Verbose output

QUICK START:
  ngcli init                                    Initialize ngcli
  ngcli generate mysite                         Interactive template selection & parameter input
  ngcli generate api --template prod            Use specific template
  ngcli list                                    List all configurations
  ngcli enable mysite                           Enable configuration
  ngcli template create custom-api              Create custom template

EXAMPLES:
  # Interactive workflow (recommended for beginners)
  ngcli generate mysite
  
  # Direct template usage
  ngcli generate blog --template dev --set domain=blog.local --set root_path=/var/www/blog
  
  # Template management
  ngcli template list
  ngcli template create api-server --from prod
  ngcli template edit api-server
  
  # Configuration management
  ngcli list
  ngcli enable blog
  ngcli disable blog --no-reload
  ngcli reload --test

For detailed help on any command, use:
  ngcli help [command]
  ngcli [command] --help`)
}

func showInitHelp() {
	fmt.Println(`Initialize ngcli with default templates and configuration

USAGE:
  ngcli init [flags]

DESCRIPTION:
  Creates the template directory and copies sample template configurations
  for different environments (prod, staging, dev). Also checks permissions
  for nginx configuration directories.

EXAMPLES:
  ngcli init                    Initialize with default settings
  ngcli init --verbose          Initialize with verbose output

The init command will:
  1. Create ~/.ngcli/templates/ directory
  2. Copy sample templates (prod.conf.tpl, staging.conf.tpl, dev.conf.tpl)
  3. Check write permissions for nginx directories`)
}

func showGenerateHelp() {
	fmt.Println(`Generate nginx configuration from template

USAGE:
  ngcli generate <config_name> [flags]

FLAGS:
  -t, --template string   Template to use (if not specified, shows available templates)
      --set stringArray   Set template parameters (key=value)
  -i, --interactive      Interactive mode for parameter input
      --dry-run          Preview output without writing files
      --force            Overwrite existing files without confirmation
  -o, --output string    Override output file path

DESCRIPTION:
  Generates nginx configuration file from a template with specified parameters.
  The config_name becomes the output filename (config_name.conf).
  
  If no template is specified, shows available templates to choose from.
  If no parameters are provided, automatically prompts for interactive input.

WORKFLOW OPTIONS:

  1. Interactive Mode (Recommended):
     ngcli generate mysite
     # Shows template selection menu
     # Guides through parameter input
     
  2. Direct Template Usage:
     ngcli generate api-server --template prod --set domain=api.example.com
     
  3. Interactive Parameters with Specific Template:
     ngcli generate blog --template dev --interactive
     
  4. Preview Mode:
     ngcli generate test --template staging --dry-run

EXAMPLES:
  # Interactive workflow
  ngcli generate mysite
  # Output: /etc/nginx/sites-available/mysite.conf
  
  # Direct usage
  ngcli generate api --template prod --set domain=api.example.com --set ssl_cert=/path/to/cert --set ssl_key=/path/to/key --set root_path=/var/www/api
  
  # Custom output location
  ngcli generate temp --template dev --set domain=temp.local --output /tmp/nginx-temp.conf
  
  # Preview without creating file
  ngcli generate preview --template staging --set domain=staging.example.com --dry-run

TEMPLATE PARAMETER SYSTEM:
  Templates use comment-based metadata for parameter definitions:
  
  # @param domain string required "Primary domain"
  # @param port integer optional "Server port" default=80
  
  Use 'ngcli template show <template_name> --params' to see parameter details.`)
}

func showListHelp() {
	fmt.Println(`List nginx configurations or templates

USAGE:
  ngcli list [flags]

FLAGS:
  -t, --templates   List available templates instead of configurations

DESCRIPTION:
  Lists nginx configuration files in the output directory or available 
  templates in the template directory. Shows status (enabled/disabled) 
  for configurations on Debian/Ubuntu systems.

EXAMPLES:
  ngcli list              List all nginx configurations
  ngcli list --templates  List available templates
  ngcli list -t           List available templates (short form)`)
}

func showShowHelp() {
	fmt.Println(`Show contents of nginx configuration file

USAGE:
  ngcli show <config_name>

DESCRIPTION:
  Displays the contents of a nginx configuration file. The config name 
  should be specified without the .conf extension for sites-available 
  configurations.

EXAMPLES:
  ngcli show default      Show contents of default configuration
  ngcli show mysite       Show contents of mysite configuration`)
}

func showEnableHelp() {
	fmt.Println(`Enable nginx configuration by creating symlink

USAGE:
  ngcli enable <config_name> [flags]

FLAGS:
  --no-reload   Skip automatic nginx reload

DESCRIPTION:
  Enables nginx configuration by creating a symbolic link in sites-enabled 
  directory (Debian/Ubuntu systems only). Automatically reloads nginx 
  configuration to apply changes unless --no-reload flag is used.
  
  This command makes the configuration active immediately.

EXAMPLES:
  ngcli enable mysite              Enable configuration and reload nginx
  ngcli enable api-server          Enable configuration and reload nginx
  ngcli enable blog --no-reload    Enable configuration without reloading
  
NOTES:
  - Only works on Debian/Ubuntu systems with sites-enabled directory
  - Requires write permissions to /etc/nginx/sites-enabled
  - Automatically creates backup if configuration already enabled`)
}

func showDisableHelp() {
	fmt.Println(`Disable nginx configuration by removing symlink

USAGE:
  ngcli disable <config_name> [flags]

FLAGS:
  --no-reload   Skip automatic nginx reload

DESCRIPTION:
  Disables nginx configuration by removing the symbolic link from 
  sites-enabled directory (Debian/Ubuntu systems only). Automatically 
  reloads nginx configuration to apply changes unless --no-reload flag is used.
  
  This command makes the configuration inactive immediately.

EXAMPLES:
  ngcli disable mysite              Disable configuration and reload nginx
  ngcli disable api-server          Disable configuration and reload nginx  
  ngcli disable blog --no-reload    Disable configuration without reloading
  
NOTES:
  - Only works on Debian/Ubuntu systems with sites-enabled directory
  - Does not delete the original configuration file
  - Use 'ngcli delete' to remove the configuration file entirely`)
}

func showDeleteHelp() {
	fmt.Println(`Delete nginx configuration file

USAGE:
  ngcli delete <config_name> [flags]

FLAGS:
  --force   Force deletion without confirmation

DESCRIPTION:
  Deletes nginx configuration file and removes any associated symlink.
  Prompts for confirmation unless --force flag is used.

EXAMPLES:
  ngcli delete mysite         Delete configuration with confirmation
  ngcli delete mysite --force Delete configuration without confirmation`)
}

func showReloadHelp() {
	fmt.Println(`Reload nginx configuration

USAGE:
  ngcli reload [flags]

FLAGS:
  --dry-run        Preview command without execution
  -t, --test       Test configuration syntax before reloading

DESCRIPTION:
  Reloads nginx configuration to apply changes. Can test configuration 
  syntax before reloading with --test flag.

EXAMPLES:
  ngcli reload              Reload nginx configuration
  ngcli reload --test       Test configuration then reload
  ngcli reload --dry-run    Preview reload command without executing`)
}

func showTemplateHelp() {
	fmt.Println(`Manage nginx configuration templates

USAGE:
  ngcli template <subcommand> [args] [flags]

AVAILABLE SUBCOMMANDS:
  create      Create a new template
  list        List all available templates  
  show        Show template content and metadata
  edit        Edit template in text editor
  delete      Delete a custom template
  validate    Validate template syntax

EXAMPLES:
  # Template creation
  ngcli template create api-server              Create new template from scratch
  ngcli template create my-blog --from prod     Clone from existing template
  
  # Template viewing
  ngcli template list                           List all templates with descriptions
  ngcli template show prod                      Show template content and parameters
  ngcli template show prod --params             Show only parameter information
  
  # Template editing
  ngcli template edit api-server                Edit with auto-detected editor
  ngcli template edit api-server --editor nano  Edit with specific editor
  
  # Template management
  ngcli template validate api-server            Check template syntax
  ngcli template delete api-server              Delete custom template

TEMPLATE METADATA FORMAT:
  Templates use comment-based metadata for parameter definitions:
  
  # Template: api-server
  # Description: API server with SSL and rate limiting
  # Author: your-name
  # Version: 1.0
  #
  # @param domain string required "Primary domain"
  # @param port integer optional "Server port" default=3000
  # @param ssl_cert file_path required "SSL certificate path"
  
EDITOR SELECTION:
  Editor priority: --editor flag → $VISUAL → $EDITOR → system default
  
NOTES:
  - Built-in templates (prod, staging, dev) cannot be deleted
  - Custom templates are stored in ~/.ngcli/templates/
  - Templates must have .conf.tpl extension`)
}