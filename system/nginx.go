package system

import (
	"fmt"
	"os/exec"
)

func NginxReload() error {
	cmd := exec.Command("nginx", "-s", "reload")
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to reload nginx: %w", err)
	}
	
	return nil
}

func NginxTest() error {
	cmd := exec.Command("nginx", "-t")
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("nginx configuration test failed: %s", string(output))
	}
	
	return nil
}

func NginxStatus() error {
	cmd := exec.Command("nginx", "-v")
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("nginx is not available: %w", err)
	}
	
	return nil
}