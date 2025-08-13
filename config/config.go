package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Config struct {
	TemplateDir string            `yaml:"template_dir"`
	OutputDir   string            `yaml:"output_dir"`
	Verbose     bool              `yaml:"verbose"`
	Defaults    map[string]string `yaml:"defaults"`
}

func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		TemplateDir: filepath.Join(homeDir, ".ngcli", "templates"),
		OutputDir:   "",
		Verbose:     false,
		Defaults: map[string]string{
			"root_path": "/var/www/html",
			"ssl_cert":  "/etc/ssl/certs/nginx.crt",
			"ssl_key":   "/etc/ssl/private/nginx.key",
		},
	}
}

func Load() (*Config, error) {
	configPath := getConfigPath()
	
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	
	return &config, nil
}

func (c *Config) Save() error {
	configPath := getConfigPath()
	
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}

func getConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".ngcli", "config.yaml")
}

func (c *Config) MergeDefaults(params map[string]string) map[string]string {
	merged := make(map[string]string)
	
	for key, value := range c.Defaults {
		merged[key] = value
	}
	
	for key, value := range params {
		merged[key] = value
	}
	
	return merged
}