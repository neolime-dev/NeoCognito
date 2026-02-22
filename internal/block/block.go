// Package block defines the core Block data type and its Markdown/YAML parser.
package block

import (
	"bufio"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

var wikilinkRegex = regexp.MustCompile(`\[\[([^\]]+)\]\]`)

// Status constants for task-like blocks.
const (
	StatusNone     = ""
	StatusTodo     = "todo"
	StatusDoing    = "doing"
	StatusDone     = "done"
	StatusArchived = "archived"
)

// Block is the atomic unit of NeoCognito — everything is a Block.
// A Block with no Status behaves as a note/wiki entry.
// A Block with a Status behaves as a task.
type Block struct {
	ID          string     `yaml:"id"`
	Title       string     `yaml:"title"`
	Status      string     `yaml:"status,omitempty"`
	Tags        []string   `yaml:"tags,omitempty"`
	Due         *time.Time `yaml:"due,omitempty"`
	Links       []string   `yaml:"links,omitempty"`
	Recur       string     `yaml:"recur,omitempty"`        // v0.3: recurrence rule
	DelegatedTo string     `yaml:"delegated_to,omitempty"` // v0.3: handoff
	Pomodoros   int        `yaml:"pomodoros,omitempty"`    // v0.3: focus sessions
	TimeSpent   int        `yaml:"time_spent,omitempty"`   // v0.3: time spent in minutes
	Area        string     `yaml:"area,omitempty"`         // v0.4: PARA area
	Source      string     `yaml:"source,omitempty"`       // v0.5: URL/origin
	Created     time.Time  `yaml:"created"`
	Modified    time.Time  `yaml:"modified"`
	Body        string     `yaml:"-"`
	FilePath    string     `yaml:"-"`
}

// NextStatus cycles the block status: "" -> todo -> doing -> done -> archived -> todo.
func (b *Block) NextStatus() {
	switch b.Status {
	case StatusNone:
		b.Status = StatusTodo
	case StatusTodo:
		b.Status = StatusDoing
	case StatusDoing:
		b.Status = StatusDone
	case StatusDone:
		b.Status = StatusArchived
	case StatusArchived:
		b.Status = StatusTodo
	}
}

// IsTask returns true if the block has a task status.
func (b *Block) IsTask() bool {
	return b.Status != StatusNone
}

// BodyPreview returns the first n characters of the body for search snippets.
func (b *Block) BodyPreview(n int) string {
	body := strings.TrimSpace(b.Body)
	if len(body) <= n {
		return body
	}
	return body[:n]
}

// GenerateID creates a short random hex ID for a new block.
func GenerateID() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		// Fallback to time-based generation if crypto/rand fails
		return fmt.Sprintf("%x", time.Now().UnixNano()&0xffffffff)
	}
	return fmt.Sprintf("%x", b)
}

// NewBlock creates a new Block with a generated ID and timestamps.
func NewBlock(title string) *Block {
	now := time.Now()
	return &Block{
		ID:       GenerateID(),
		Title:    title,
		Created:  now,
		Modified: now,
	}
}

// ParseFile reads a Markdown file and returns a Block.
// It expects optional YAML frontmatter delimited by "---" lines.
func ParseFile(path string) (*Block, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading block file: %w", err)
	}

	block, err := Parse(string(data))
	if err != nil {
		return nil, fmt.Errorf("parsing block file %s: %w", path, err)
	}
	block.FilePath = path

	// If ID is missing, generate a stable one from the RELATIVE filepath
	if block.ID == "" {
		// Use filename as a base for human-readable-ish stable IDs
		base := filepath.Base(path)
		base = strings.TrimSuffix(base, ".md")

		// Add a hash of the full path for uniqueness
		h := sha1.New()
		h.Write([]byte(path))
		hashStr := fmt.Sprintf("%x", h.Sum(nil))[:6]

		block.ID = base + "-" + hashStr
	}

	return block, nil
}

// Parse parses a string containing Markdown with optional YAML frontmatter.
func Parse(content string) (*Block, error) {
	block := &Block{}

	frontmatter, body, hasFrontmatter := splitFrontmatter(content)
	if hasFrontmatter {
		if err := yaml.Unmarshal([]byte(frontmatter), block); err != nil {
			return nil, fmt.Errorf("parsing frontmatter: %w", err)
		}
	}

	block.Body = body

	// Default ID if missing
	if block.ID == "" {
		block.ID = GenerateID()
	}
	// Default timestamps
	if block.Created.IsZero() {
		block.Created = time.Now()
	}
	if block.Modified.IsZero() {
		block.Modified = block.Created
	}

	// Extract wikilinks from body
	matches := wikilinkRegex.FindAllStringSubmatch(block.Body, -1)
	for _, match := range matches {
		if len(match) > 1 {
			link := strings.TrimSpace(match[1])
			// Avoid duplicates
			exists := false
			for _, l := range block.Links {
				if l == link {
					exists = true
					break
				}
			}
			if !exists {
				block.Links = append(block.Links, link)
			}
		}
	}

	return block, nil
}

// Marshal serializes a Block back into Markdown with YAML frontmatter.
func Marshal(b *Block) (string, error) {
	b.Modified = time.Now()

	fm, err := yaml.Marshal(b)
	if err != nil {
		return "", fmt.Errorf("marshaling frontmatter: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("---\n")
	sb.Write(fm)
	sb.WriteString("---\n")
	if b.Body != "" {
		sb.WriteString("\n")
		sb.WriteString(b.Body)
		if !strings.HasSuffix(b.Body, "\n") {
			sb.WriteString("\n")
		}
	}
	return sb.String(), nil
}

// WriteFile writes a Block to its FilePath.
func WriteFile(b *Block) error {
	content, err := Marshal(b)
	if err != nil {
		return err
	}
	return os.WriteFile(b.FilePath, []byte(content), 0644)
}

// WriteFileVersioned writes a Block to its FilePath and saves the previous
// version to versionsDir/<id>/<timestamp>.md before overwriting.
func WriteFileVersioned(b *Block, versionsDir string) error {
	// Snapshot current file if it exists
	if _, err := os.Stat(b.FilePath); err == nil {
		snapshotDir := filepath.Join(versionsDir, b.ID)
		if err := os.MkdirAll(snapshotDir, 0755); err == nil {
			ts := time.Now().Format("2006-01-02T15-04-05")
			snap := filepath.Join(snapshotDir, ts+".md")
			if data, err := os.ReadFile(b.FilePath); err == nil {
				if writeErr := os.WriteFile(snap, data, 0644); writeErr != nil {
					// Log it or return? Returning could prevent saving. We should probably return to avoid data loss.
					return fmt.Errorf("failed to create version snapshot: %w", writeErr)
				}
			}
		}
	}
	return WriteFile(b)
}

// IsStale returns true if the block hasn't been modified in more than staleDays
// and is not completed or archived.
func (b *Block) IsStale(staleDays int) bool {
	if b.Status == StatusDone || b.Status == StatusArchived {
		return false
	}
	return time.Since(b.Modified) > time.Duration(staleDays)*24*time.Hour
}

// splitFrontmatter separates YAML frontmatter from the Markdown body.
// Returns (frontmatter, body, hasFrontmatter).
func splitFrontmatter(content string) (string, string, bool) {
	scanner := bufio.NewScanner(strings.NewReader(content))

	// First line must be "---"
	if !scanner.Scan() || strings.TrimSpace(scanner.Text()) != "---" {
		return "", content, false
	}

	var fmLines []string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			// Found closing delimiter — rest is body
			var bodyLines []string
			for scanner.Scan() {
				bodyLines = append(bodyLines, scanner.Text())
			}
			body := strings.Join(bodyLines, "\n")
			return strings.Join(fmLines, "\n"), strings.TrimPrefix(body, "\n"), true
		}
		fmLines = append(fmLines, line)
	}

	// No closing "---" found — treat everything as body
	return "", content, false
}
