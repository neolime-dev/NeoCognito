package sync

// engine_test.go is in the same package so it can test unexported helpers
// like sanitizeFilename and reach internal fields directly.

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	gosync "sync"
	"testing"
	"time"

	"github.com/neolime-dev/neocognito/internal/block"
	"github.com/neolime-dev/neocognito/internal/config"
	"github.com/neolime-dev/neocognito/internal/store"
)

// --- stubStore implements store.Storer in-memory for test isolation. ---

type stubStore struct {
	mu      gosync.Mutex
	blocks  map[string]*block.Block
	upserts []string // IDs of upserted blocks, in order
	deletes []string // IDs of deleted blocks
}

func newStub() *stubStore {
	return &stubStore{blocks: make(map[string]*block.Block)}
}

func (s *stubStore) Close() error { return nil }
func (s *stubStore) UpsertBlock(b *block.Block) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.blocks[b.ID] = b
	s.upserts = append(s.upserts, b.ID)
	return nil
}
func (s *stubStore) DeleteBlock(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.blocks, id)
	s.deletes = append(s.deletes, id)
	return nil
}
func (s *stubStore) GetBlock(id string) (*block.Block, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, ok := s.blocks[id]
	if !ok {
		return nil, nil
	}
	return b, nil
}
func (s *stubStore) GetBlockByPath(fp string) (*block.Block, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, b := range s.blocks {
		if b.FilePath == fp {
			return b, nil
		}
	}
	return nil, nil
}
func (s *stubStore) GetBlockByTitle(title string) (*block.Block, error) { return nil, nil }
func (s *stubStore) ListBlocks(filter store.Filter) ([]*block.Block, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []*block.Block
	for _, b := range s.blocks {
		out = append(out, b)
	}
	return out, nil
}
func (s *stubStore) Search(query string) ([]*block.Block, error) { return nil, nil }
func (s *stubStore) GetTags() ([]string, error)                  { return nil, nil }
func (s *stubStore) GetTagCounts() (map[string]int, error)       { return nil, nil }
func (s *stubStore) GetLinksFrom(blockID string) ([]string, error) { return nil, nil }
func (s *stubStore) GetLinksTo(blockID string) ([]string, error)   { return nil, nil }
func (s *stubStore) GetGraphEdges() ([]store.GraphEdge, error)     { return nil, nil }
func (s *stubStore) FindRelatedBlocks(target *block.Block, limit int) ([]*block.Block, error) {
	return nil, nil
}
func (s *stubStore) BlockCount() (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.blocks), nil
}

// compile-time check
var _ store.Storer = (*stubStore)(nil)

// --- helpers ---

func newEngine(t *testing.T, st store.Storer, cfg *config.Config) *Engine {
	t.Helper()
	dir := t.TempDir()
	blocksDir := filepath.Join(dir, "blocks")
	if err := os.MkdirAll(blocksDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	e := NewEngine(blocksDir, st, cfg)
	// Silence log output in tests.
	e.SetLogger(slog.New(slog.NewTextHandler(io.Discard, nil)))
	return e
}

func writeMD(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeMD %s: %v", name, err)
	}
	return path
}

const minimalFrontmatter = "---\nid: \"testid\"\ntitle: \"Test\"\n---\n\nBody text.\n"

// --- FullScan ---

func TestFullScan_IndexesMarkdownFiles(t *testing.T) {
	st := newStub()
	e := newEngine(t, st, nil)

	// 3 md files and 1 txt — only md should be indexed
	for _, name := range []string{"a.md", "b.md", "c.md"} {
		writeMD(t, e.blocksDir, name, minimalFrontmatter)
	}
	writeMD(t, e.blocksDir, "ignore.txt", "not markdown")

	if err := e.FullScan(); err != nil {
		t.Fatalf("FullScan: %v", err)
	}

	st.mu.Lock()
	count := len(st.upserts)
	st.mu.Unlock()

	// Each file may resolve the same frontmatter ID, but we get one upsert per file walk
	if count != 3 {
		t.Errorf("expected 3 upserts, got %d", count)
	}
}

func TestFullScan_IgnoresParseErrors(t *testing.T) {
	st := newStub()
	e := newEngine(t, st, nil)

	writeMD(t, e.blocksDir, "valid.md", minimalFrontmatter)
	// empty file — will fail parsing (no content)
	writeMD(t, e.blocksDir, "empty.md", "")

	// FullScan should not return an error just because one file fails
	if err := e.FullScan(); err != nil {
		t.Fatalf("FullScan returned error on bad file: %v", err)
	}

	st.mu.Lock()
	count := len(st.upserts)
	st.mu.Unlock()

	// valid.md should be indexed; empty.md may or may not be (ParseFile on empty
	// returns a block with no ID but that's not a fatal error for FullScan).
	if count < 1 {
		t.Errorf("expected at least 1 upsert, got %d", count)
	}
}

// --- CreateBlock ---

func TestCreateBlock_BasicFields(t *testing.T) {
	st := newStub()
	e := newEngine(t, st, nil)

	b, err := e.CreateBlock("Test Note")
	if err != nil {
		t.Fatalf("CreateBlock: %v", err)
	}
	if b == nil {
		t.Fatal("returned nil block")
	}
	if b.ID == "" {
		t.Error("ID should be non-empty")
	}
	if b.Title != "Test Note" {
		t.Errorf("title: want %q, got %q", "Test Note", b.Title)
	}
	// File should exist on disk
	if _, err := os.Stat(b.FilePath); os.IsNotExist(err) {
		t.Errorf("block file not created: %s", b.FilePath)
	}
	// Should be inside the blocks dir
	if !filepath.HasPrefix(b.FilePath, e.blocksDir) {
		t.Errorf("FilePath %q not under blocksDir %q", b.FilePath, e.blocksDir)
	}
}

func TestCreateBlock_NLDateExtraction(t *testing.T) {
	st := newStub()
	e := newEngine(t, st, nil)

	b, err := e.CreateBlock("Buy groceries tomorrow")
	if err != nil {
		t.Fatalf("CreateBlock: %v", err)
	}
	if b.Due == nil {
		t.Fatal("expected Due to be set from NL date")
	}
	if b.Title != "Buy groceries" {
		t.Errorf("title after NL extraction: want %q, got %q", "Buy groceries", b.Title)
	}
	// Due should be tomorrow
	want := time.Now().AddDate(0, 0, 1)
	if b.Due.Year() != want.Year() || b.Due.Month() != want.Month() || b.Due.Day() != want.Day() {
		t.Errorf("due date: want %s, got %s", want.Format("2006-01-02"), b.Due.Format("2006-01-02"))
	}
}

func TestCreateBlock_DefaultTagsApplied(t *testing.T) {
	st := newStub()
	cfg := &config.Config{DefaultTags: []string{"@inbox"}}
	e := newEngine(t, st, cfg)

	b, err := e.CreateBlock("Tagged block")
	if err != nil {
		t.Fatalf("CreateBlock: %v", err)
	}
	found := false
	for _, tag := range b.Tags {
		if tag == "@inbox" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("default tag @inbox not applied; tags: %v", b.Tags)
	}
}

func TestCreateBlock_DeduplicatesFilenames(t *testing.T) {
	st := newStub()
	e := newEngine(t, st, nil)

	b1, err := e.CreateBlock("Fix bug")
	if err != nil {
		t.Fatalf("CreateBlock 1: %v", err)
	}
	b2, err := e.CreateBlock("Fix bug")
	if err != nil {
		t.Fatalf("CreateBlock 2: %v", err)
	}

	if b1.FilePath == b2.FilePath {
		t.Errorf("both blocks got the same FilePath: %s", b1.FilePath)
	}
	for _, fp := range []string{b1.FilePath, b2.FilePath} {
		if _, err := os.Stat(fp); os.IsNotExist(err) {
			t.Errorf("file not found: %s", fp)
		}
	}
}

// --- UpdateBlock ---

func TestUpdateBlock_WritesFileToDisk(t *testing.T) {
	st := newStub()
	e := newEngine(t, st, nil)

	b, err := e.CreateBlock("Update me")
	if err != nil {
		t.Fatalf("CreateBlock: %v", err)
	}

	b.Body = "## Updated body\n\nNew content here.\n"
	if err := e.UpdateBlock(b); err != nil {
		t.Fatalf("UpdateBlock: %v", err)
	}

	data, err := os.ReadFile(b.FilePath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !containsStr(string(data), "Updated body") {
		t.Errorf("file does not contain updated body: %s", data)
	}
}

// --- DeleteBlock ---

func TestDeleteBlock_RemovesFile(t *testing.T) {
	st := newStub()
	e := newEngine(t, st, nil)

	b, err := e.CreateBlock("Delete me")
	if err != nil {
		t.Fatalf("CreateBlock: %v", err)
	}
	fp := b.FilePath

	// Seed stub so DeleteBlock finds the record
	_ = st.UpsertBlock(b)

	if err := e.DeleteBlock(b); err != nil {
		t.Fatalf("DeleteBlock: %v", err)
	}

	if _, err := os.Stat(fp); !os.IsNotExist(err) {
		t.Errorf("file still exists after DeleteBlock: %s", fp)
	}

	st.mu.Lock()
	deleted := st.deletes
	st.mu.Unlock()
	if len(deleted) == 0 {
		t.Error("store.DeleteBlock was not called")
	}
}

// --- sanitizeFilename (unexported, accessible from same package) ---

func TestSanitizeFilename(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"Hello World", "hello-world"},
		{"Multiple   spaces", "multiple---spaces"},
		{"café", "caf"},                      // non-ASCII stripped
		{"", ""},                              // empty
		{"123-abc_DEF", "123-abc_def"},        // digits, hyphens, underscores kept
		{"!@#$%", ""},                         // all special → empty
		{longStr(80), longStr(64)[:64]},       // truncated to 64
	}
	for _, tc := range cases {
		got := sanitizeFilename(tc.input)
		if got != tc.want {
			t.Errorf("sanitizeFilename(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func longStr(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'a'
	}
	return string(b)
}

func containsStr(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (haystack == needle ||
		func() bool {
			for i := 0; i <= len(haystack)-len(needle); i++ {
				if haystack[i:i+len(needle)] == needle {
					return true
				}
			}
			return false
		}())
}
