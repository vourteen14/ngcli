package template

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type Template struct {
	Name     string
	Path     string
	Content  string
	Template *template.Template
	Metadata *TemplateMetadata
}

func LoadTemplate(name, templateDir string) (*Template, error) {
	templateName := name
	if !strings.HasSuffix(name, ".conf.tpl") {
		templateName = name + ".conf.tpl"
	}
	
	templatePath := filepath.Join(templateDir, templateName)
	
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("template not found: %s", templatePath)
	}
	
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template %s: %w", templatePath, err)
	}
	
	tmpl, err := template.New(name).Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", templatePath, err)
	}
	
	metadata, err := ParseTemplateMetadata(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template metadata: %w", err)
	}
	
	return &Template{
		Name:     name,
		Path:     templatePath,
		Content:  string(content),
		Template: tmpl,
		Metadata: metadata,
	}, nil
}

func (t *Template) Render(params map[string]string) (string, error) {
	var output strings.Builder
	
	if err := t.Template.Execute(&output, params); err != nil {
		return "", fmt.Errorf("failed to render template %s: %w", t.Name, err)
	}
	
	return output.String(), nil
}

func (t *Template) RenderWithValidation(params map[string]string) (string, error) {
	paramsWithDefaults := t.Metadata.ApplyDefaults(params)
	
	if err := t.Metadata.ValidateParameters(paramsWithDefaults); err != nil {
		return "", err
	}
	
	return t.Render(paramsWithDefaults)
}

func ListTemplates(templateDir string) ([]string, error) {
	var templates []string
	
	entries, err := os.ReadDir(templateDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read template directory %s: %w", templateDir, err)
	}
	
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		name := entry.Name()
		if strings.HasSuffix(name, ".conf.tpl") {
			templateName := strings.TrimSuffix(name, ".conf.tpl")
			templates = append(templates, templateName)
		}
	}
	
	return templates, nil
}

func ValidateTemplate(templatePath string) error {
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}
	
	_, err = template.New("validate").Parse(string(content))
	if err != nil {
		return fmt.Errorf("template syntax error: %w", err)
	}
	
	return nil
}