// Package assets manages binary file attachments for blocks.
package assets

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Manager handles binary asset storage.
type Manager struct {
	assetsDir string
}

// NewManager creates an asset manager rooted at the given directory.
func NewManager(assetsDir string) *Manager {
	return &Manager{assetsDir: assetsDir}
}

// Store copies a file into the managed assets directory and returns
// the relative path suitable for embedding in a block's Markdown.
func (m *Manager) Store(srcPath string) (string, error) {
	if err := os.MkdirAll(m.assetsDir, 0755); err != nil {
		return "", fmt.Errorf("creating assets dir: %w", err)
	}

	ext := filepath.Ext(srcPath)
	base := strings.TrimSuffix(filepath.Base(srcPath), ext)

	// Generate unique name
	destName := fmt.Sprintf("%s-%d%s", sanitize(base), time.Now().UnixNano(), ext)
	destPath := filepath.Join(m.assetsDir, destName)

	src, err := os.Open(srcPath)
	if err != nil {
		return "", fmt.Errorf("opening source: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("creating destination: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("copying file: %w", err)
	}

	// Return relative path for Markdown embedding
	return filepath.Join("assets", destName), nil
}

// Delete removes an asset file from the managed directory.
func (m *Manager) Delete(relativePath string) error {
	fullPath := filepath.Join(filepath.Dir(m.assetsDir), relativePath)
	return os.Remove(fullPath)
}

func sanitize(name string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		case r == ' ':
			b.WriteRune('-')
		}
	}
	result := b.String()
	if len(result) > 32 {
		result = result[:32]
	}
	return result
}
