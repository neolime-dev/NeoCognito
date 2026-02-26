package store_test

import (
	"database/sql"
	"path/filepath"
	gosync "sync"
	"testing"
	"time"

	"github.com/neolime-dev/neocognito/internal/block"
	"github.com/neolime-dev/neocognito/internal/store"
)

// newTestStore opens a fresh Store backed by a temp-dir file DB.
// It is closed automatically when the test ends.
func newTestStore(t *testing.T) store.Storer {
	t.Helper()
	dir := t.TempDir()
	st, err := store.New(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	return st
}

// newBlock returns a minimal valid Block with a unique FilePath under dir.
func newBlock(dir string) *block.Block {
	b := block.NewBlock("Test Block")
	b.FilePath = filepath.Join(dir, b.ID+".md")
	return b
}

// --- Migration ---

func TestMigrate_IdempotentOnReopen(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "idem.db")

	st1, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("first open: %v", err)
	}
	st1.Close()

	st2, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("second open: %v", err)
	}
	st2.Close()
}

// --- CRUD ---

func TestUpsertBlock_InsertAndRetrieve(t *testing.T) {
	st := newTestStore(t)
	dir := t.TempDir()

	b := newBlock(dir)
	b.Title = "Hello World"
	b.Status = block.StatusTodo
	b.Area = "work"
	b.Tags = []string{"@test"}
	due := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	b.Due = &due

	if err := st.UpsertBlock(b); err != nil {
		t.Fatalf("UpsertBlock: %v", err)
	}

	got, err := st.GetBlock(b.ID)
	if err != nil {
		t.Fatalf("GetBlock: %v", err)
	}
	if got.Title != "Hello World" {
		t.Errorf("title: want %q got %q", "Hello World", got.Title)
	}
	if got.Status != block.StatusTodo {
		t.Errorf("status: want %q got %q", block.StatusTodo, got.Status)
	}
	if got.Area != "work" {
		t.Errorf("area: want %q got %q", "work", got.Area)
	}
	if got.Due == nil {
		t.Fatal("due: expected non-nil")
	}
	if !got.Due.Truncate(time.Second).Equal(due) {
		t.Errorf("due: want %v got %v", due, *got.Due)
	}
}

func TestUpsertBlock_UpdateExisting(t *testing.T) {
	st := newTestStore(t)
	dir := t.TempDir()

	b := newBlock(dir)
	b.Title = "Original"
	if err := st.UpsertBlock(b); err != nil {
		t.Fatalf("initial upsert: %v", err)
	}

	b.Title = "Updated"
	if err := st.UpsertBlock(b); err != nil {
		t.Fatalf("update upsert: %v", err)
	}

	got, err := st.GetBlock(b.ID)
	if err != nil {
		t.Fatalf("GetBlock: %v", err)
	}
	if got.Title != "Updated" {
		t.Errorf("expected updated title, got %q", got.Title)
	}
}

func TestUpsertBlock_TagsReplacedOnUpdate(t *testing.T) {
	st := newTestStore(t)
	dir := t.TempDir()

	b := newBlock(dir)
	b.Tags = []string{"a", "b"}
	if err := st.UpsertBlock(b); err != nil {
		t.Fatalf("first upsert: %v", err)
	}

	b.Tags = []string{"c"}
	if err := st.UpsertBlock(b); err != nil {
		t.Fatalf("second upsert: %v", err)
	}

	tags, err := st.GetTags()
	if err != nil {
		t.Fatalf("GetTags: %v", err)
	}
	if len(tags) != 1 || tags[0] != "c" {
		t.Errorf("expected tags=[c], got %v", tags)
	}
}

func TestDeleteBlock_RemovesFromIndex(t *testing.T) {
	st := newTestStore(t)
	dir := t.TempDir()

	b := newBlock(dir)
	if err := st.UpsertBlock(b); err != nil {
		t.Fatalf("UpsertBlock: %v", err)
	}
	if err := st.DeleteBlock(b.ID); err != nil {
		t.Fatalf("DeleteBlock: %v", err)
	}

	_, err := st.GetBlock(b.ID)
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows after delete, got %v", err)
	}
}

// --- ListBlocks filters ---

func TestListBlocks_FilterByStatus(t *testing.T) {
	st := newTestStore(t)
	dir := t.TempDir()

	todo := newBlock(dir)
	todo.Title = "Todo item"
	todo.Status = block.StatusTodo

	done := newBlock(dir)
	done.Title = "Done item"
	done.Status = block.StatusDone

	noStatus := newBlock(dir)
	noStatus.Title = "No status"
	noStatus.Status = ""

	for _, b := range []*block.Block{todo, done, noStatus} {
		if err := st.UpsertBlock(b); err != nil {
			t.Fatalf("UpsertBlock %q: %v", b.Title, err)
		}
	}

	todoStr := block.StatusTodo
	results, err := st.ListBlocks(store.Filter{Status: &todoStr})
	if err != nil {
		t.Fatalf("ListBlocks: %v", err)
	}
	if len(results) != 1 || results[0].Title != "Todo item" {
		t.Errorf("expected [Todo item], got %v", titlesOf(results))
	}
}

func TestListBlocks_FilterByTag(t *testing.T) {
	st := newTestStore(t)
	dir := t.TempDir()

	tagged := newBlock(dir)
	tagged.Title = "Tagged"
	tagged.Tags = []string{"@work"}

	untagged := newBlock(dir)
	untagged.Title = "Untagged"

	for _, b := range []*block.Block{tagged, untagged} {
		if err := st.UpsertBlock(b); err != nil {
			t.Fatalf("UpsertBlock: %v", err)
		}
	}

	results, err := st.ListBlocks(store.Filter{Tag: "@work"})
	if err != nil {
		t.Fatalf("ListBlocks: %v", err)
	}
	if len(results) != 1 || results[0].Title != "Tagged" {
		t.Errorf("expected [Tagged], got %v", titlesOf(results))
	}
}

func TestListBlocks_FilterByDueBefore(t *testing.T) {
	st := newTestStore(t)
	dir := t.TempDir()

	soon := newBlock(dir)
	soon.Title = "Due soon"
	nearDue := time.Now().Add(2 * 24 * time.Hour)
	soon.Due = &nearDue

	far := newBlock(dir)
	far.Title = "Due far"
	farDue := time.Now().Add(60 * 24 * time.Hour)
	far.Due = &farDue

	for _, b := range []*block.Block{soon, far} {
		if err := st.UpsertBlock(b); err != nil {
			t.Fatalf("UpsertBlock: %v", err)
		}
	}

	cutoff := time.Now().Add(7 * 24 * time.Hour)
	results, err := st.ListBlocks(store.Filter{DueBefore: &cutoff})
	if err != nil {
		t.Fatalf("ListBlocks: %v", err)
	}
	if len(results) != 1 || results[0].Title != "Due soon" {
		t.Errorf("expected [Due soon], got %v", titlesOf(results))
	}
}

// --- Search ---

func TestSearch_FTS5MatchesTitle(t *testing.T) {
	st := newTestStore(t)
	dir := t.TempDir()

	b := newBlock(dir)
	b.Title = "Quantum mechanics notes"
	b.Body = "Some notes about quantum physics."
	if err := st.UpsertBlock(b); err != nil {
		t.Fatalf("UpsertBlock: %v", err)
	}

	results, err := st.Search("quantum")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one search result")
	}
	found := false
	for _, r := range results {
		if r.ID == b.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("upserted block not in search results: %v", titlesOf(results))
	}
}

func TestSearch_EmptyQuery(t *testing.T) {
	st := newTestStore(t)

	results, err := st.Search("")
	if err != nil {
		t.Fatalf("Search(\"\") error: %v", err)
	}
	if results != nil {
		t.Errorf("expected nil for empty query, got %v", results)
	}
}

// --- Tags ---

func TestGetTagCounts(t *testing.T) {
	st := newTestStore(t)
	dir := t.TempDir()

	b1 := newBlock(dir)
	b1.Tags = []string{"project", "inbox"}
	b2 := newBlock(dir)
	b2.Tags = []string{"project"}
	b3 := newBlock(dir)
	b3.Tags = []string{"inbox"}

	for _, b := range []*block.Block{b1, b2, b3} {
		if err := st.UpsertBlock(b); err != nil {
			t.Fatalf("UpsertBlock: %v", err)
		}
	}

	counts, err := st.GetTagCounts()
	if err != nil {
		t.Fatalf("GetTagCounts: %v", err)
	}
	if counts["project"] != 2 {
		t.Errorf("project count: want 2, got %d", counts["project"])
	}
	if counts["inbox"] != 2 {
		t.Errorf("inbox count: want 2, got %d", counts["inbox"])
	}
}

// --- Links ---

func TestGetLinksFrom_And_GetLinksTo(t *testing.T) {
	st := newTestStore(t)
	dir := t.TempDir()

	// Upsert B first so link resolution by ID works.
	blockB := newBlock(dir)
	blockB.Title = "Block B"
	if err := st.UpsertBlock(blockB); err != nil {
		t.Fatalf("UpsertBlock B: %v", err)
	}

	blockA := newBlock(dir)
	blockA.Title = "Block A"
	blockA.Links = []string{blockB.ID}
	if err := st.UpsertBlock(blockA); err != nil {
		t.Fatalf("UpsertBlock A: %v", err)
	}

	linksFrom, err := st.GetLinksFrom(blockA.ID)
	if err != nil {
		t.Fatalf("GetLinksFrom: %v", err)
	}
	if len(linksFrom) != 1 || linksFrom[0] != blockB.ID {
		t.Errorf("GetLinksFrom: want [%s], got %v", blockB.ID, linksFrom)
	}

	linksTo, err := st.GetLinksTo(blockB.ID)
	if err != nil {
		t.Fatalf("GetLinksTo: %v", err)
	}
	if len(linksTo) != 1 || linksTo[0] != blockA.ID {
		t.Errorf("GetLinksTo: want [%s], got %v", blockA.ID, linksTo)
	}
}

// --- BlockCount ---

func TestBlockCount(t *testing.T) {
	st := newTestStore(t)
	dir := t.TempDir()

	for i := 0; i < 3; i++ {
		b := newBlock(dir)
		if err := st.UpsertBlock(b); err != nil {
			t.Fatalf("UpsertBlock %d: %v", i, err)
		}
	}

	count, err := st.BlockCount()
	if err != nil {
		t.Fatalf("BlockCount: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3, got %d", count)
	}
}

// --- Concurrency (regression for Task 15 WAL pool change) ---

func TestConcurrentReads(t *testing.T) {
	st := newTestStore(t)
	dir := t.TempDir()

	for i := 0; i < 5; i++ {
		b := newBlock(dir)
		if err := st.UpsertBlock(b); err != nil {
			t.Fatalf("seed UpsertBlock: %v", err)
		}
	}

	var wg gosync.WaitGroup
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := st.ListBlocks(store.Filter{}); err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Errorf("concurrent read error: %v", err)
	}
}

// titlesOf extracts block titles for readable error messages.
func titlesOf(blocks []*block.Block) []string {
	titles := make([]string, len(blocks))
	for i, b := range blocks {
		titles[i] = b.Title
	}
	return titles
}
