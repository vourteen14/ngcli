package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func WriteFile(path, content string, force bool) error {
	if _, err := os.Stat(path); err == nil && !force {
		return fmt.Errorf("file already exists: %s (use --force to overwrite)", path)
	}
	
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}
	
	return nil
}

func BackupFile(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	
	timestamp := time.Now().Format("20060102-150405")
	backupPath := fmt.Sprintf("%s.backup-%s", path, timestamp)
	
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read original file %s: %w", path, err)
	}
	
	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		return fmt.Errorf("failed to create backup %s: %w", backupPath, err)
	}
	
	return nil
}

func ReadFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", path, err)
	}
	
	return string(content), nil
}

func ListConfigs(dir string) ([]string, error) {
	var configs []string
	
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory does not exist: %s", dir)
	}
	
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}
	
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		name := entry.Name()
		if strings.HasSuffix(name, ".conf") || !strings.Contains(name, ".") {
			configs = append(configs, name)
		}
	}
	
	return configs, nil
}

func CreateSymlink(src, dst string) error {
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return fmt.Errorf("source file does not exist: %s", src)
	}
	
	if _, err := os.Lstat(dst); err == nil {
		if err := os.Remove(dst); err != nil {
			return fmt.Errorf("failed to remove existing symlink %s: %w", dst, err)
		}
	}
	
	if err := os.Symlink(src, dst); err != nil {
		return fmt.Errorf("failed to create symlink %s -> %s: %w", dst, src, err)
	}
	
	return nil
}

func RemoveSymlink(path string) error {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to stat %s: %w", path, err)
	}
	
	if info.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("%s is not a symbolic link", path)
	}
	
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to remove symlink %s: %w", path, err)
	}
	
	return nil
}

func DeleteFile(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", path)
	}
	
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete file %s: %w", path, err)
	}
	
	return nil
}

func CheckWritePermission(dir string) error {
	testFile := filepath.Join(dir, ".ngcli-write-test")
	
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("no write permission to directory %s: %w", dir, err)
	}
	
	os.Remove(testFile)
	
	return nil
}