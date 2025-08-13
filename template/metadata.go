package template

import (
	"bufio"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type TemplateMetadata struct {
	Name        string
	Description string
	Author      string
	Version     string
	Parameters  []ParameterInfo
}

type ParameterInfo struct {
	Name        string
	Type        string
	Required    bool
	Description string
	Default     string
	Options     []string
}

func ParseTemplateMetadata(templateContent string) (*TemplateMetadata, error) {
	scanner := bufio.NewScanner(strings.NewReader(templateContent))
	metadata := &TemplateMetadata{
		Parameters: make([]ParameterInfo, 0),
	}
	
	templateLineRegex := regexp.MustCompile(`^#\s*Template:\s*(.+)$`)
	descriptionRegex := regexp.MustCompile(`^#\s*Description:\s*(.+)$`)
	authorRegex := regexp.MustCompile(`^#\s*Author:\s*(.+)$`)
	versionRegex := regexp.MustCompile(`^#\s*Version:\s*(.+)$`)
	paramRegex := regexp.MustCompile(`^#\s*@param\s+(\w+)\s+(\w+)\s+(required|optional)\s+"([^"]+)"(?:\s+(.*))?$`)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		if !strings.HasPrefix(line, "#") && line != "" {
			break
		}
		
		if match := templateLineRegex.FindStringSubmatch(line); match != nil {
			metadata.Name = strings.TrimSpace(match[1])
			continue
		}
		
		if match := descriptionRegex.FindStringSubmatch(line); match != nil {
			metadata.Description = strings.TrimSpace(match[1])
			continue
		}
		
		if match := authorRegex.FindStringSubmatch(line); match != nil {
			metadata.Author = strings.TrimSpace(match[1])
			continue
		}
		
		if match := versionRegex.FindStringSubmatch(line); match != nil {
			metadata.Version = strings.TrimSpace(match[1])
			continue
		}
		
		if match := paramRegex.FindStringSubmatch(line); match != nil {
			param := ParameterInfo{
				Name:        match[1],
				Type:        match[2],
				Required:    match[3] == "required",
				Description: match[4],
			}
			
			if len(match) > 5 && match[5] != "" {
				attributes := match[5]
				param.Default, param.Options = parseParameterAttributes(attributes)
			}
			
			metadata.Parameters = append(metadata.Parameters, param)
		}
	}
	
	return metadata, nil
}

func parseParameterAttributes(attributes string) (string, []string) {
	var defaultValue string
	var options []string
	
	defaultRegex := regexp.MustCompile(`default=([^,\s]+|"[^"]*")`)
	if match := defaultRegex.FindStringSubmatch(attributes); match != nil {
		defaultValue = strings.Trim(match[1], `"`)
	}
	
	optionsRegex := regexp.MustCompile(`options=\[([^\]]+)\]`)
	if match := optionsRegex.FindStringSubmatch(attributes); match != nil {
		optionsList := match[1]
		for _, opt := range strings.Split(optionsList, ",") {
			cleaned := strings.Trim(strings.TrimSpace(opt), `"`)
			if cleaned != "" {
				options = append(options, cleaned)
			}
		}
	}
	
	return defaultValue, options
}

func (m *TemplateMetadata) ValidateParameters(params map[string]string) error {
	var missing []string
	var invalid []string
	
	for _, param := range m.Parameters {
		if param.Required {
			if _, exists := params[param.Name]; !exists {
				missing = append(missing, param.Name)
			}
		}
		
		if value, exists := params[param.Name]; exists {
			if err := m.validateParameterValue(param, value); err != nil {
				invalid = append(invalid, fmt.Sprintf("%s: %v", param.Name, err))
			}
		}
	}
	
	if len(missing) > 0 {
		return fmt.Errorf("missing required parameters: %s", strings.Join(missing, ", "))
	}
	
	if len(invalid) > 0 {
		return fmt.Errorf("invalid parameter values: %s", strings.Join(invalid, "; "))
	}
	
	return nil
}

func (m *TemplateMetadata) validateParameterValue(param ParameterInfo, value string) error {
	switch param.Type {
	case "integer":
		if _, err := strconv.Atoi(value); err != nil {
			return fmt.Errorf("must be an integer")
		}
	case "boolean":
		if value != "true" && value != "false" {
			return fmt.Errorf("must be true or false")
		}
	case "file_path":
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("file path cannot be empty")
		}
	}
	
	if len(param.Options) > 0 {
		for _, option := range param.Options {
			if value == option {
				return nil 
			}
		}
		return fmt.Errorf("must be one of: %s", strings.Join(param.Options, ", "))
	}
	
	return nil
}

// GetParameterHelp returns formatted help text for parameters
func (m *TemplateMetadata) GetParameterHelp() string {
	if len(m.Parameters) == 0 {
		return "No parameters defined for this template"
	}
	
	var help strings.Builder
	
	help.WriteString("Parameters:\n")
	for _, param := range m.Parameters {
		required := "optional"
		if param.Required {
			required = "required"
		}
		
		help.WriteString(fmt.Sprintf("  %-15s %-8s %-8s %s\n", 
			param.Name, param.Type, required, param.Description))
		
		if param.Default != "" {
			help.WriteString(fmt.Sprintf("  %-15s default: %s\n", "", param.Default))
		}
		
		if len(param.Options) > 0 {
			help.WriteString(fmt.Sprintf("  %-15s options: %s\n", "", strings.Join(param.Options, ", ")))
		}
		
		help.WriteString("\n")
	}
	
	return help.String()
}

// ApplyDefaults applies default values to parameters if not provided
func (m *TemplateMetadata) ApplyDefaults(params map[string]string) map[string]string {
	result := make(map[string]string)
	
	// Copy existing parameters
	for key, value := range params {
		result[key] = value
	}
	
	// Apply defaults for missing parameters
	for _, param := range m.Parameters {
		if _, exists := result[param.Name]; !exists && param.Default != "" {
			result[param.Name] = param.Default
		}
	}
	
	return result
}