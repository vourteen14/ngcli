# ngcli

A command-line tool for managing Nginx configurations using templates.

## Features

- Generate Nginx configurations from templates with parameter validation
- Interactive template selection and parameter input
- Template management (create, edit, list, validate)
- Automatic Nginx reload after enabling/disabling configurations
- Cross-platform support (Debian/Ubuntu and RedHat/CentOS)
- Comment-based parameter definitions in templates

## Installation

```bash
git clone https://github.com/vourteen14/ngcli.git
cd ngcli
go build -o ngcli
sudo mv ngcli /usr/local/bin/
```

## Quick Start

```bash
# Initialize with default templates
ngcli init

# Generate configuration interactively
ngcli generate mysite

# List configurations
ngcli list

# Enable configuration
ngcli enable mysite

# Reload nginx
ngcli reload
```

## Usage

### Configuration Management

```bash
# Generate configuration with template selection
ngcli generate mysite

# Generate with specific template
ngcli generate api --template prod --set domain=api.example.com

# List all configurations
ngcli list

# Show configuration content
ngcli show mysite

# Enable/disable configurations
ngcli enable mysite
ngcli disable mysite

# Delete configuration
ngcli delete mysite
```

### Template Management

```bash
# List available templates
ngcli template list

# Create new template
ngcli template create api-server

# Create template from existing one
ngcli template create my-blog --from prod

# Edit template
ngcli template edit api-server

# Show template details
ngcli template show prod --params

# Validate template syntax
ngcli template validate api-server
```

## Template System

Templates use comment-based metadata for parameter definitions:

```nginx
# Template: api-server
# Description: API server with SSL and rate limiting
# Author: your-name
# Version: 1.0
#
# @param domain string required "Primary domain"
# @param port integer optional "Server port" default=3000
# @param ssl_cert file_path required "SSL certificate path"

server {
    listen {{.port}} ssl http2;
    server_name {{.domain}};
    
    ssl_certificate {{.ssl_cert}};
    
    location / {
        proxy_pass http://localhost:{{.port}};
    }
}
```

### Parameter Types

- `string` - Text value
- `integer` - Numeric value
- `boolean` - true/false value
- `file_path` - File system path
- `array` - Multiple values

### Parameter Attributes

- `required` - Parameter must be provided
- `optional` - Parameter is optional
- `default=value` - Default value if not specified
- `options=["opt1","opt2"]` - Allowed values

## Directory Structure

```
~/.ngcli/
├── templates/
│   ├── prod.conf.tpl
│   ├── staging.conf.tpl
│   ├── dev.conf.tpl
│   └── custom-templates.conf.tpl
```

Generated configurations are placed in:
- `/etc/nginx/sites-available/` (Debian/Ubuntu)
- `/etc/nginx/conf.d/` (RedHat/CentOS)

## Commands

| Command | Description |
|---------|-------------|
| `init` | Initialize ngcli with default templates |
| `generate` | Generate configuration from template |
| `list` | List configurations or templates |
| `show` | Show configuration content |
| `enable` | Enable configuration (create symlink + reload) |
| `disable` | Disable configuration (remove symlink + reload) |
| `delete` | Delete configuration file |
| `reload` | Reload nginx configuration |
| `template` | Manage templates |

## Global Flags

- `--template-dir` - Override template directory
- `--output-dir` - Override output directory
- `--verbose` - Enable verbose output

## Examples

### Basic Web Server

```bash
ngcli generate blog --template dev \
  --set domain=blog.local \
  --set root_path=/var/www/blog
```

### Production SSL Site

```bash
ngcli generate mysite --template prod \
  --set domain=example.com \
  --set ssl_cert=/etc/ssl/certs/example.com.crt \
  --set ssl_key=/etc/ssl/private/example.com.key \
  --set root_path=/var/www/html
```

### API Server

```bash
ngcli template create api --from prod
ngcli template edit api
ngcli generate myapi --template api \
  --set domain=api.example.com \
  --set port=3000
```

## Requirements

- Go 1.21+
- Nginx
- Linux (Debian/Ubuntu or RedHat/CentOS)
- Root privileges for nginx operations

## Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/new-feature`)
3. Commit changes (`git commit -am 'Add new feature'`)
4. Push to branch (`git push origin feature/new-feature`)
5. Create Pull Request

## License

MIT License - see [LICENSE](LICENSE) file for details.
