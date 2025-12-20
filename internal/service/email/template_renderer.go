package email

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	texttemplate "text/template"
)

type TemplateData struct {
	AppName        string
	RecipientEmail string
	ActionURL      string
	ActionLabel    string
	ExpiresAt      string
	SupportEmail   string
}

type RenderedTemplate struct {
	Subject string
	Text    string
	HTML    string
}

type Renderer struct {
	baseDir string
}

func NewRenderer(baseDir string) *Renderer {
	trimmed := strings.TrimSpace(baseDir)
	if trimmed == "" {
		trimmed = "templates/email"
	}
	return &Renderer{baseDir: trimmed}
}

func (r *Renderer) Render(name string, data TemplateData) (RenderedTemplate, error) {
	if r == nil {
		return RenderedTemplate{}, errors.New("email templates: renderer is nil")
	}
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return RenderedTemplate{}, errors.New("email templates: name is empty")
	}

	subjectPath := filepath.Join(r.baseDir, trimmed+".subject.tmpl")
	textPath := filepath.Join(r.baseDir, trimmed+".txt")
	htmlPath := filepath.Join(r.baseDir, trimmed+".html")

	subject, err := renderTextTemplate(subjectPath, data, true)
	if err != nil {
		return RenderedTemplate{}, err
	}
	textBody, err := renderTextTemplate(textPath, data, false)
	if err != nil {
		return RenderedTemplate{}, err
	}
	htmlBody, err := renderHTMLTemplate(htmlPath, data, false)
	if err != nil {
		return RenderedTemplate{}, err
	}

	return RenderedTemplate{
		Subject: strings.TrimSpace(subject),
		Text:    textBody,
		HTML:    htmlBody,
	}, nil
}

func renderTextTemplate(path string, data TemplateData, required bool) (string, error) {
	if !required {
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				return "", nil
			}
			return "", err
		}
	} else if err := ensureFile(path); err != nil {
		return "", err
	}

	tmpl, err := texttemplate.ParseFiles(path)
	if err != nil {
		return "", fmt.Errorf("email templates: parse %s: %w", path, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("email templates: execute %s: %w", path, err)
	}

	return buf.String(), nil
}

func renderHTMLTemplate(path string, data TemplateData, required bool) (string, error) {
	if !required {
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				return "", nil
			}
			return "", err
		}
	} else if err := ensureFile(path); err != nil {
		return "", err
	}

	tmpl, err := template.ParseFiles(path)
	if err != nil {
		return "", fmt.Errorf("email templates: parse %s: %w", path, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("email templates: execute %s: %w", path, err)
	}

	return buf.String(), nil
}

func ensureFile(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("email templates: file not found: %s", path)
		}
		return err
	}
	return nil
}
