package utils

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

func ParseSetFlags(setFlags []string) (map[string]string, error) {
	params := make(map[string]string)
	
	for _, flag := range setFlags {
		parts := strings.SplitN(flag, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid set flag format: %s (expected key=value)", flag)
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		if key == "" {
			return nil, fmt.Errorf("empty key in set flag: %s", flag)
		}
		
		params[key] = value
	}
	
	return params, nil
}

func DetectNginxConfigPath() (string, error) {
	sitesAvailable := "/etc/nginx/sites-available"
	if _, err := os.Stat(sitesAvailable); err == nil {
		return sitesAvailable, nil
	}
	
	confD := "/etc/nginx/conf.d"
	if _, err := os.Stat(confD); err == nil {
		return confD, nil
	}
	
	if runtime.GOOS == "linux" {
		if isDebianBased() {
			return sitesAvailable, nil
		}
		return confD, nil
	}
	
	return "", fmt.Errorf("unable to detect nginx configuration directory")
}

func DetectNginxEnabledPath() (string, bool) {
	sitesEnabled := "/etc/nginx/sites-enabled"
	if _, err := os.Stat(sitesEnabled); err == nil {
		return sitesEnabled, true
	}
	return "", false
}

func isDebianBased() bool {
	debianFiles := []string{
		"/etc/debian_version",
		"/etc/lsb-release",
	}
	
	for _, file := range debianFiles {
		if _, err := os.Stat(file); err == nil {
			return true
		}
	}
	
	return false
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func ValidateRequiredParams(params map[string]string, required []string) error {
	var missing []string
	
	for _, param := range required {
		if _, exists := params[param]; !exists {
			missing = append(missing, param)
		}
	}
	
	if len(missing) > 0 {
		return fmt.Errorf("missing required parameters: %s", strings.Join(missing, ", "))
	}
	
	return nil
}