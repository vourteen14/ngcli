package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/your-username/ngcli/filesystem"
	"github.com/your-username/ngcli/utils"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize ngcli with default templates and configuration",
	Long: `Initialize ngcli by creating the template directory and copying
sample template configurations for different environments.

It will also check permissions for nginx configuration directories.`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	fmt.Println("Initializing ngcli configuration")

	if err := createTemplateDirectory(); err != nil {
		return fmt.Errorf("failed to create template directory: %w", err)
	}

	if err := createSampleTemplates(); err != nil {
		return fmt.Errorf("failed to create sample templates: %w", err)
	}

	if err := checkNginxPermissions(); err != nil {
		fmt.Printf("Warning: %v\n", err)
		fmt.Println("Administrative privileges may be required for nginx configuration operations")
	}

	fmt.Println("ngcli initialization completed successfully")
	fmt.Printf("Template directory: %s\n", templateDir)
	fmt.Println("Use 'ngcli generate' to create nginx configurations")

	return nil
}

func createTemplateDirectory() error {
	if err := utils.EnsureDir(templateDir); err != nil {
		return err
	}

	if verbose {
		fmt.Printf("Created template directory: %s\n", templateDir)
	}

	return nil
}

func createSampleTemplates() error {
	templates := map[string]string{
		"prod.conf.tpl":    prodTemplate,
		"staging.conf.tpl": stagingTemplate,
		"dev.conf.tpl":     devTemplate,
	}

	for filename, content := range templates {
		filePath := filepath.Join(templateDir, filename)
		
		if utils.FileExists(filePath) {
			if verbose {
				fmt.Printf("Template already exists: %s\n", filename)
			}
			continue
		}

		if err := filesystem.WriteFile(filePath, content, false); err != nil {
			return fmt.Errorf("failed to create %s: %w", filename, err)
		}

		if verbose {
			fmt.Printf("Created template: %s\n", filename)
		}
	}

	return nil
}

func checkNginxPermissions() error {
	nginxDirs := []string{
		"/etc/nginx/sites-available",
		"/etc/nginx/sites-enabled",
		"/etc/nginx/conf.d",
	}

	for _, dir := range nginxDirs {
		if !utils.FileExists(dir) {
			continue
		}

		if err := filesystem.CheckWritePermission(dir); err != nil {
			return fmt.Errorf("no write permission to %s", dir)
		}

		if verbose {
			fmt.Printf("Write permission verified: %s\n", dir)
		}
	}

	return nil
}

const prodTemplate = `# Template: prod
# Description: Production-ready reverse proxy with maximum security
# Author: ngcli
# Version: 1.0
#
# @param domain string required "Primary domain for the service"
# @param upstream_host string required "Backend service host" default="127.0.0.1"
# @param upstream_port integer required "Backend service port" default=3000
# @param ssl_cert file_path required "Path to SSL certificate file"
# @param ssl_key file_path required "Path to SSL private key file"
# @param client_max_body_size string optional "Maximum request body size" default="10m"

# Rate limiting zones
limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
limit_req_zone $binary_remote_addr zone=login:10m rate=5r/m;
limit_conn_zone $binary_remote_addr zone=conn_limit_per_ip:10m;

# Security headers map
map $sent_http_content_type $nosniff_header {
    ~^text/ "nosniff";
    default "";
}

server {
    listen 80;
    server_name {{.domain}};
    
    # Security: Force HTTPS redirect
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name {{.domain}};
    
    # SSL Configuration - Production Grade
    ssl_certificate {{.ssl_cert}};
    ssl_certificate_key {{.ssl_key}};
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    ssl_session_timeout 1d;
    ssl_session_cache shared:SSL:50m;
    ssl_stapling on;
    ssl_stapling_verify on;
    
    # Security Headers - Production Grade
    add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload" always;
    add_header X-Frame-Options "DENY" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline';" always;
    add_header Permissions-Policy "geolocation=(), microphone=(), camera=()" always;
    
    # Connection and rate limits
    limit_conn conn_limit_per_ip 20;
    limit_req zone=api burst=20 nodelay;
    
    # Basic security settings
    client_max_body_size {{.client_max_body_size}};
    server_tokens off;
    
    # Hide nginx version
    more_clear_headers Server;
    
    # Security: Block common attack patterns
    location ~* /(\.git|\.svn|\.env|config\.json|package\.json) {
        deny all;
        return 404;
    }
    
    # Main proxy configuration
    location / {
        proxy_pass http://{{.upstream_host}}:{{.upstream_port}};
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Forwarded-Host $host;
        proxy_set_header X-Forwarded-Port $server_port;
        
        # Timeouts
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
        
        # Buffer settings
        proxy_buffering on;
        proxy_buffer_size 8k;
        proxy_buffers 8 8k;
        
        # Security: Remove potentially dangerous headers
        proxy_hide_header X-Powered-By;
        proxy_hide_header Server;
    }
    
    # Rate limited login endpoint
    location /login {
        limit_req zone=login burst=3 nodelay;
        proxy_pass http://{{.upstream_host}}:{{.upstream_port}};
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    # Health check endpoint (internal only)
    location /health {
        access_log off;
        proxy_pass http://{{.upstream_host}}:{{.upstream_port}};
        allow 127.0.0.1;
        allow 10.0.0.0/8;
        allow 172.16.0.0/12;
        allow 192.168.0.0/16;
        deny all;
    }
    
    # Static assets with caching
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        proxy_pass http://{{.upstream_host}}:{{.upstream_port}};
        proxy_cache_valid 200 302 1h;
        proxy_cache_valid 404 1m;
        add_header Cache-Control "public, immutable";
        expires 1y;
    }
}`

const stagingTemplate = `# Template: staging
# Description: Staging environment with basic security and debugging capabilities
# Author: ngcli
# Version: 1.0
#
# @param domain string required "Staging domain"
# @param upstream_host string required "Backend service host" default="127.0.0.1"
# @param upstream_port integer required "Backend service port" default=3000
# @param auth_file file_path optional "Basic auth file path" default="/etc/nginx/.htpasswd"
# @param ssl_enabled string optional "Enable SSL" default="no" options=["yes","no"]
# @param ssl_cert file_path optional "Path to SSL certificate file"
# @param ssl_key file_path optional "Path to SSL private key file"

# Rate limiting for staging (more lenient)
limit_req_zone $binary_remote_addr zone=staging_api:10m rate=30r/s;

server {
    listen 80;
    server_name {{.domain}};
    
    # Basic auth for staging access
    {{if .auth_file}}auth_basic "Staging Environment - Authorized Access Only";
    auth_basic_user_file {{.auth_file}};{{end}}
    
    # Basic security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header X-Environment "staging" always;
    
    # Development-friendly settings
    add_header X-Debug-Backend "{{.upstream_host}}:{{.upstream_port}}" always;
    add_header X-Request-ID "$request_id" always;
    
    # Rate limiting (lenient)
    limit_req zone=staging_api burst=50 nodelay;
    
    client_max_body_size 50m;
    
    # Main proxy configuration
    location / {
        proxy_pass http://{{.upstream_host}}:{{.upstream_port}};
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Request-ID $request_id;
        
        # Generous timeouts for debugging
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
        
        # Debug headers
        add_header X-Upstream-Response-Time $upstream_response_time always;
        add_header X-Upstream-Status $upstream_status always;
    }
    
    # Health and debug endpoints
    location /health {
        proxy_pass http://{{.upstream_host}}:{{.upstream_port}};
        access_log off;
    }
    
    location /debug {
        proxy_pass http://{{.upstream_host}}:{{.upstream_port}};
        proxy_set_header X-Debug-Mode "enabled";
    }
    
    # Enhanced logging for staging
    access_log /var/log/nginx/{{.domain}}_access.log combined;
    error_log /var/log/nginx/{{.domain}}_error.log info;
}

# SSL server block (conditional)
{{if eq .ssl_enabled "yes"}}
server {
    listen 443 ssl http2;
    server_name {{.domain}};
    
    ssl_certificate {{.ssl_cert}};
    ssl_certificate_key {{.ssl_key}};
    ssl_protocols TLSv1.2 TLSv1.3;
    
    # Same configuration as HTTP block above
    {{if .auth_file}}auth_basic "Staging Environment - Authorized Access Only";
    auth_basic_user_file {{.auth_file}};{{end}}
    
    add_header X-Environment "staging-ssl" always;
    
    location / {
        proxy_pass http://{{.upstream_host}}:{{.upstream_port}};
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
{{end}}`

const devTemplate = `# Template: dev
# Description: Development environment with minimal security and maximum debugging
# Author: ngcli
# Version: 1.0
#
# @param domain string required "Development domain" default="dev.local"
# @param upstream_host string required "Backend service host" default="127.0.0.1"
# @param upstream_port integer required "Backend service port" default=3000
# @param debug_mode string optional "Enable debug mode" default="on" options=["on","off"]

server {
    listen 80;
    server_name {{.domain}};
    
    # Development-friendly settings
    client_max_body_size 100m;
    
    # CORS headers for local development
    add_header Access-Control-Allow-Origin "*" always;
    add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS, PATCH" always;
    add_header Access-Control-Allow-Headers "DNT,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization,X-API-Key" always;
    add_header Access-Control-Expose-Headers "Content-Length,Content-Range,X-Request-ID" always;
    
    # Debug headers
    add_header X-Environment "development" always;
    add_header X-Debug-Mode "{{.debug_mode}}" always;
    add_header X-Backend "{{.upstream_host}}:{{.upstream_port}}" always;
    add_header X-Request-ID "$request_id" always;
    add_header X-Response-Time "$upstream_response_time" always;
    
    # Handle preflight requests
    location ~ ^/.*$ {
        if ($request_method = 'OPTIONS') {
            add_header Access-Control-Allow-Origin "*";
            add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS, PATCH";
            add_header Access-Control-Allow-Headers "DNT,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization,X-API-Key";
            add_header Access-Control-Max-Age 1728000;
            add_header Content-Type "text/plain charset=UTF-8";
            add_header Content-Length 0;
            return 204;
        }
    }
    
    # Main proxy configuration
    location / {
        proxy_pass http://{{.upstream_host}}:{{.upstream_port}};
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Request-ID $request_id;
        proxy_set_header X-Dev-Mode "true";
        
        # Very generous timeouts for debugging
        proxy_connect_timeout 300s;
        proxy_send_timeout 300s;
        proxy_read_timeout 300s;
        
        # Disable buffering for real-time debugging
        proxy_buffering off;
        proxy_request_buffering off;
        
        # Debug response headers
        add_header X-Upstream-Response-Time "$upstream_response_time" always;
        add_header X-Upstream-Status "$upstream_status" always;
        add_header X-Upstream-Address "$upstream_addr" always;
    }
    
    # Development tools endpoints
    location /dev-tools {
        proxy_pass http://{{.upstream_host}}:{{.upstream_port}};
        proxy_set_header X-Dev-Tools "enabled";
    }
    
    location /metrics {
        proxy_pass http://{{.upstream_host}}:{{.upstream_port}};
        add_header X-Metrics-Access "dev-mode" always;
    }
    
    # Hot reload support for development servers
    location /hot-reload {
        proxy_pass http://{{.upstream_host}}:{{.upstream_port}};
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
    
    # Verbose logging for development
    access_log /var/log/nginx/{{.domain}}_access.log combined;
    error_log /var/log/nginx/{{.domain}}_error.log debug;
}`