// Package template handles loading and parsing of Markdown templates.
package template

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/neolime-dev/neocognito/internal/block"
)

// Template represents a loaded boilerplate block.
type Template struct {
	Name    string
	Block   *block.Block
	RawText string
}

// Manager handles template discovery and loading.
type Manager struct {
	dir string
}

// NewManager creates a new template manager for the given directory.
// It creates the default templates if the directory doesn't exist.
func NewManager(dir string) (*Manager, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating template dir: %w", err)
	}

	m := &Manager{dir: dir}
	if err := m.ensureDefaults(); err != nil {
		return nil, err
	}
	return m, nil
}

// List returns all available templates.
func (m *Manager) List() ([]Template, error) {
	entries, err := os.ReadDir(m.dir)
	if err != nil {
		return nil, fmt.Errorf("reading template dir: %w", err)
	}

	var parsed []Template
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		path := filepath.Join(m.dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		b, err := block.Parse(string(data))
		if err != nil {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".md")
		if b.Title == "" {
			b.Title = name
		}

		parsed = append(parsed, Template{
			Name:    name,
			Block:   b,
			RawText: string(data),
		})
	}

	return parsed, nil
}

// Get finds a template by exact name (e.g. "meeting" for "meeting.md").
func (m *Manager) Get(name string) (Template, error) {
	// Defend against path traversal
	cleanName := filepath.Clean(name)
	if strings.Contains(cleanName, "/") || strings.Contains(cleanName, "\\") {
		return Template{}, fmt.Errorf("invalid template name: %s", name)
	}

	path := filepath.Join(m.dir, cleanName+".md")
	data, err := os.ReadFile(path)
	if err != nil {
		return Template{}, fmt.Errorf("template not found: %s", name)
	}

	b, err := block.Parse(string(data))
	if err != nil {
		return Template{}, fmt.Errorf("parsing template %s: %w", name, err)
	}

	if b.Title == "" {
		b.Title = cleanName
	}
	return Template{
		Name:    cleanName,
		Block:   b,
		RawText: string(data),
	}, nil
}

// Instantiate creates a new block object based on the template.
func (t Template) Instantiate() *block.Block {
	// We re-parse from RawText to get a fresh Block object with new ID/timestamps
	b, _ := block.Parse(t.RawText)

	// Reset runtime metadata
	b.ID = block.GenerateID()
	now := time.Now()
	b.Created = now
	b.Modified = now

	// Variables expansion: replace {{date}} and {{time}} if present
	dateStr := now.Format("2006-01-02")
	timeStr := now.Format("15:04")
	b.Title = strings.ReplaceAll(b.Title, "{{date}}", dateStr)
	b.Title = strings.ReplaceAll(b.Title, "{{time}}", timeStr)
	b.Body = strings.ReplaceAll(b.Body, "{{date}}", dateStr)
	b.Body = strings.ReplaceAll(b.Body, "{{time}}", timeStr)

	return b
}

func (m *Manager) ensureDefaults() error {
	defaults := map[string]string{
		"meeting.md": `---
title: Meeting with {{name}} — {{date}}
tags:
  - "@meeting"
---
## 👥 Attendees
- 

## 📝 Notes
- 

## ✅ Action Items
- [ ] 
`,
		"project.md": `---
title: Project: 
status: todo
tags:
  - "@project"
area: projects
---
## 🎯 Objectives
1. 

## 🔗 Resources
- 

## 📋 Tasks
- [ ] 
`,
		"book-note.md": `---
title: "Book: "
tags:
  - "@reading"
  - "@book"
---
**Author**: 
**Finished**: {{date}}
**Rating**: /5

## 💡 Key Insights
1. 

## 📝 Summary

`,
	}

	for name, content := range defaults {
		path := filepath.Join(m.dir, name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				return fmt.Errorf("writing default template %s: %w", name, err)
			}
		}
	}
	return nil
}
