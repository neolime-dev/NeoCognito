package export_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neolime-dev/neocognito/internal/block"
	"github.com/neolime-dev/neocognito/internal/export"
	"github.com/neolime-dev/neocognito/internal/store"
)

// --- stub store ---

type stubStore struct {
	blocks []*block.Block
}

func (s *stubStore) Close() error                                          { return nil }
func (s *stubStore) UpsertBlock(b *block.Block) error                     { return nil }
func (s *stubStore) DeleteBlock(id string) error                           { return nil }
func (s *stubStore) GetBlock(id string) (*block.Block, error)              { return nil, nil }
func (s *stubStore) GetBlockByPath(fp string) (*block.Block, error)        { return nil, nil }
func (s *stubStore) GetBlockByTitle(title string) (*block.Block, error)    { return nil, nil }
func (s *stubStore) ListBlocks(filter store.Filter) ([]*block.Block, error) {
	return s.blocks, nil
}
func (s *stubStore) Search(query string) ([]*block.Block, error)                { return nil, nil }
func (s *stubStore) GetTags() ([]string, error)                                  { return nil, nil }
func (s *stubStore) GetTagCounts() (map[string]int, error)                       { return nil, nil }
func (s *stubStore) GetLinksFrom(blockID string) ([]string, error)               { return nil, nil }
func (s *stubStore) GetLinksTo(blockID string) ([]string, error)                 { return nil, nil }
func (s *stubStore) GetGraphEdges() ([]store.GraphEdge, error)                   { return nil, nil }
func (s *stubStore) FindRelatedBlocks(b *block.Block, n int) ([]*block.Block, error) {
	return nil, nil
}
func (s *stubStore) BlockCount() (int, error) { return len(s.blocks), nil }

// compile-time check
var _ store.Storer = (*stubStore)(nil)

// --- helpers ---

// makeBlock writes a real .md file and returns a block pointing to it.
func makeBlock(t *testing.T, dir, title, body string) *block.Block {
	t.Helper()
	b := block.NewBlock(title)
	b.Body = body
	b.FilePath = filepath.Join(dir, b.ID+".md")
	if err := block.WriteFile(b); err != nil {
		t.Fatalf("WriteFile %q: %v", title, err)
	}
	return b
}

// --- tests ---

func TestRun_CreatesIndexHTML(t *testing.T) {
	filesDir := t.TempDir()
	b1 := makeBlock(t, filesDir, "First Note", "Some content")
	b2 := makeBlock(t, filesDir, "Second Note", "More content")

	st := &stubStore{blocks: []*block.Block{b1, b2}}
	outDir := filepath.Join(t.TempDir(), "export")

	if err := export.Run(st, outDir); err != nil {
		t.Fatalf("Run: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "index.html"))
	if err != nil {
		t.Fatalf("index.html not created: %v", err)
	}
	html := string(data)
	if !strings.Contains(html, "First Note") {
		t.Error("index.html missing 'First Note'")
	}
	if !strings.Contains(html, "Second Note") {
		t.Error("index.html missing 'Second Note'")
	}
}

func TestRun_CreatesBlockHTMLFiles(t *testing.T) {
	filesDir := t.TempDir()
	b1 := makeBlock(t, filesDir, "Alpha", "")
	b2 := makeBlock(t, filesDir, "Beta", "")

	st := &stubStore{blocks: []*block.Block{b1, b2}}
	outDir := t.TempDir()

	if err := export.Run(st, outDir); err != nil {
		t.Fatalf("Run: %v", err)
	}

	for _, b := range []*block.Block{b1, b2} {
		fp := filepath.Join(outDir, b.ID+".html")
		if _, err := os.Stat(fp); os.IsNotExist(err) {
			t.Errorf("expected %s.html to exist", b.ID)
		}
	}
}

func TestRun_BlockHTMLContainsRenderedMarkdown(t *testing.T) {
	filesDir := t.TempDir()
	b := makeBlock(t, filesDir, "MD Block", "## Heading\n\nSome **bold** text.\n")
	st := &stubStore{blocks: []*block.Block{b}}
	outDir := t.TempDir()

	if err := export.Run(st, outDir); err != nil {
		t.Fatalf("Run: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outDir, b.ID+".html"))
	if err != nil {
		t.Fatalf("block HTML not found: %v", err)
	}
	html := string(data)
	if !strings.Contains(html, "<h2>") {
		t.Error("expected <h2> tag from goldmark, not found")
	}
	if !strings.Contains(html, "<strong>") {
		t.Error("expected <strong> tag from goldmark, not found")
	}
}

func TestRun_OutDirCreatedIfMissing(t *testing.T) {
	outDir := filepath.Join(t.TempDir(), "nested", "export")
	st := &stubStore{}

	if err := export.Run(st, outDir); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if _, err := os.Stat(outDir); os.IsNotExist(err) {
		t.Error("outDir was not created")
	}
}

func TestRun_EmptyStore(t *testing.T) {
	st := &stubStore{}
	outDir := t.TempDir()

	if err := export.Run(st, outDir); err != nil {
		t.Fatalf("Run with empty store: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "index.html")); os.IsNotExist(err) {
		t.Error("index.html should exist even for empty store")
	}
}

func TestRun_SkipsMissingFile(t *testing.T) {
	// Block points to a non-existent file — Run should continue, not crash.
	b := block.NewBlock("Ghost Block")
	b.FilePath = "/nonexistent/path/ghost.md"
	st := &stubStore{blocks: []*block.Block{b}}
	outDir := t.TempDir()

	if err := export.Run(st, outDir); err != nil {
		t.Fatalf("Run with missing file should not error, got: %v", err)
	}
	// index.html must still be created
	if _, err := os.Stat(filepath.Join(outDir, "index.html")); os.IsNotExist(err) {
		t.Error("index.html not created despite skipped block")
	}
}
