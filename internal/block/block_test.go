package block

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseWithFrontmatter(t *testing.T) {
	content := `---
id: "abc123"
title: "Test Block"
status: "todo"
tags: ["@work", "@urgent"]
created: "2026-02-20T15:00:00-03:00"
modified: "2026-02-20T15:30:00-03:00"
---

## Hello World

This is the body.
`

	b, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if b.ID != "abc123" {
		t.Errorf("expected ID 'abc123', got '%s'", b.ID)
	}
	if b.Title != "Test Block" {
		t.Errorf("expected Title 'Test Block', got '%s'", b.Title)
	}
	if b.Status != StatusTodo {
		t.Errorf("expected Status 'todo', got '%s'", b.Status)
	}
	if len(b.Tags) != 2 || b.Tags[0] != "@work" {
		t.Errorf("unexpected Tags: %v", b.Tags)
	}
	if !strings.Contains(b.Body, "Hello World") {
		t.Errorf("body should contain 'Hello World', got: %s", b.Body)
	}
}

func TestParseWithoutFrontmatter(t *testing.T) {
	content := "# Just a Note\n\nNo frontmatter here.\n"

	b, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if b.ID == "" {
		t.Error("expected generated ID, got empty")
	}
	if b.Title != "" {
		t.Errorf("expected empty Title, got '%s'", b.Title)
	}
	if b.Body != content {
		t.Errorf("expected full content as body, got: %s", b.Body)
	}
}

func TestMarshalRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	original := &Block{
		ID:       "round1",
		Title:    "Round Trip Test",
		Status:   StatusDoing,
		Tags:     []string{"@test"},
		Created:  now,
		Modified: now,
		Body:     "Some body content.\n",
	}

	data, err := Marshal(original)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	parsed, err := Parse(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if parsed.ID != original.ID {
		t.Errorf("ID mismatch: %s vs %s", parsed.ID, original.ID)
	}
	if parsed.Title != original.Title {
		t.Errorf("Title mismatch: %s vs %s", parsed.Title, original.Title)
	}
	if parsed.Status != original.Status {
		t.Errorf("Status mismatch: %s vs %s", parsed.Status, original.Status)
	}
	if !strings.Contains(parsed.Body, "Some body content") {
		t.Errorf("Body mismatch, got: %s", parsed.Body)
	}
}

func TestWriteAndParseFile(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "test-block.md")

	original := NewBlock("File Test")
	original.Status = StatusTodo
	original.Tags = []string{"@filetest"}
	original.Body = "Testing file write.\n"
	original.FilePath = fp

	if err := WriteFile(original); err != nil {
		t.Fatalf("write error: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		t.Fatal("file was not created")
	}

	parsed, err := ParseFile(fp)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if parsed.ID != original.ID {
		t.Errorf("ID mismatch: %s vs %s", parsed.ID, original.ID)
	}
	if parsed.FilePath != fp {
		t.Errorf("FilePath mismatch: %s vs %s", parsed.FilePath, fp)
	}
}

func TestNextStatus(t *testing.T) {
	b := &Block{}

	b.NextStatus()
	if b.Status != StatusTodo {
		t.Errorf("expected todo, got %s", b.Status)
	}
	b.NextStatus()
	if b.Status != StatusDoing {
		t.Errorf("expected doing, got %s", b.Status)
	}
	b.NextStatus()
	if b.Status != StatusDone {
		t.Errorf("expected done, got %s", b.Status)
	}
	b.NextStatus()
	if b.Status != StatusArchived {
		t.Errorf("expected archived, got %s", b.Status)
	}
	b.NextStatus()
	if b.Status != StatusTodo {
		t.Errorf("expected todo (cycle), got %s", b.Status)
	}
}

func TestBodyPreview(t *testing.T) {
	b := &Block{Body: "Hello World, this is a longer body text for testing."}

	preview := b.BodyPreview(11)
	if preview != "Hello World" {
		t.Errorf("expected 'Hello World', got '%s'", preview)
	}

	short := &Block{Body: "Short"}
	if short.BodyPreview(100) != "Short" {
		t.Errorf("expected 'Short', got '%s'", short.BodyPreview(100))
	}
}
