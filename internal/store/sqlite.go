// Package store provides SQLite-based indexing for Blocks.
package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lemondesk/neocognito/internal/block"

	_ "modernc.org/sqlite"
)

// Store wraps an SQLite database for block indexing.
type Store struct {
	db *sql.DB
}

func New(dbPath string) (*Store, error) {
	dsn := fmt.Sprintf("%s?_pragma=busy_timeout=5000&_pragma=journal_mode=WAL&_pragma=synchronous=NORMAL&_pragma=foreign_keys=ON", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// FIX: Force a single connection to serialize writes at the Go level.
	// This entirely prevents driver-level SQLite "database is locked" errors
	// caused by concurrent UI transactions and background sync loops.
	db.SetMaxOpenConns(1)

	// Performance PRAGMAs that don't need to be in DSN (though DSN is safer)
	pragmas := []string{
		"PRAGMA cache_size=-64000", // 64MB cache
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			db.Close()
			return nil, fmt.Errorf("setting pragma %q: %w", p, err)
		}
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("running migrations: %w", err)
	}
	return s, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// migrate creates the schema if it doesn't exist.
func (s *Store) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS blocks (
		id           TEXT PRIMARY KEY,
		title        TEXT NOT NULL DEFAULT '',
		status       TEXT DEFAULT '',
		due          TEXT,
		created      TEXT NOT NULL,
		modified     TEXT NOT NULL,
		filepath     TEXT NOT NULL UNIQUE,
		body_preview TEXT DEFAULT '',
		area         TEXT DEFAULT ''
	);

	CREATE TABLE IF NOT EXISTS tags (
		block_id TEXT NOT NULL,
		tag      TEXT NOT NULL,
		PRIMARY KEY (block_id, tag),
		FOREIGN KEY (block_id) REFERENCES blocks(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS links (
		source_id TEXT NOT NULL,
		target_id TEXT NOT NULL,
		PRIMARY KEY (source_id, target_id),
		FOREIGN KEY (source_id) REFERENCES blocks(id) ON DELETE CASCADE
	);
	`

	_, err := s.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("creating schema: %w", err)
	}

	// Hot migration for existing DB
	if _, err := s.db.Exec("ALTER TABLE blocks ADD COLUMN area TEXT DEFAULT ''"); err != nil {
		if !strings.Contains(err.Error(), "duplicate column name") {
			return fmt.Errorf("adding area column: %w", err)
		}
	}

	// Now create indexes
	indexes := `
	CREATE INDEX IF NOT EXISTS idx_blocks_status ON blocks(status);
	CREATE INDEX IF NOT EXISTS idx_blocks_due ON blocks(due);
	CREATE INDEX IF NOT EXISTS idx_blocks_area ON blocks(area);
	CREATE INDEX IF NOT EXISTS idx_tags_tag ON tags(tag);
	CREATE INDEX IF NOT EXISTS idx_links_target ON links(target_id);
	`
	_, err = s.db.Exec(indexes)
	if err != nil {
		return fmt.Errorf("creating indexes: %w", err)
	}

	// FTS5 virtual table for full-text search
	fts := `
	CREATE VIRTUAL TABLE IF NOT EXISTS blocks_fts USING fts5(
		id UNINDEXED,
		title,
		body_preview,
		content='blocks',
		content_rowid='rowid'
	);

	-- Triggers to keep FTS in sync
	CREATE TRIGGER IF NOT EXISTS blocks_ai AFTER INSERT ON blocks BEGIN
		INSERT INTO blocks_fts(rowid, id, title, body_preview)
		VALUES (new.rowid, new.id, new.title, new.body_preview);
	END;

	CREATE TRIGGER IF NOT EXISTS blocks_ad AFTER DELETE ON blocks BEGIN
		INSERT INTO blocks_fts(blocks_fts, rowid, id, title, body_preview)
		VALUES ('delete', old.rowid, old.id, old.title, old.body_preview);
	END;

	CREATE TRIGGER IF NOT EXISTS blocks_au AFTER UPDATE ON blocks BEGIN
		INSERT INTO blocks_fts(blocks_fts, rowid, id, title, body_preview)
		VALUES ('delete', old.rowid, old.id, old.title, old.body_preview);
		INSERT INTO blocks_fts(rowid, id, title, body_preview)
		VALUES (new.rowid, new.id, new.title, new.body_preview);
	END;
	`

	_, err = s.db.Exec(fts)
	if err != nil {
		return fmt.Errorf("creating FTS: %w", err)
	}

	return nil
}

// UpsertBlock inserts or updates a block and its tags/links in the index.
func (s *Store) UpsertBlock(b *block.Block) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	var dueStr *string
	if b.Due != nil {
		d := b.Due.Format(time.RFC3339)
		dueStr = &d
	}

	_, err = tx.Exec(`
		INSERT INTO blocks (id, title, status, due, created, modified, filepath, body_preview, area)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			title = excluded.title,
			status = excluded.status,
			due = excluded.due,
			modified = excluded.modified,
			filepath = excluded.filepath,
			body_preview = excluded.body_preview,
			area = excluded.area
		ON CONFLICT(filepath) DO UPDATE SET
			title = excluded.title,
			status = excluded.status,
			due = excluded.due,
			modified = excluded.modified,
			id = excluded.id,
			body_preview = excluded.body_preview,
			area = excluded.area
	`, b.ID, b.Title, b.Status, dueStr,
		b.Created.Format(time.RFC3339),
		b.Modified.Format(time.RFC3339),
		b.FilePath,
		b.BodyPreview(200),
		b.Area,
	)
	if err != nil {
		return fmt.Errorf("upserting block: %w", err)
	}

	// Sync tags
	if _, err := tx.Exec("DELETE FROM tags WHERE block_id = ?", b.ID); err != nil {
		return fmt.Errorf("clearing tags: %w", err)
	}
	for _, tag := range b.Tags {
		if _, err := tx.Exec("INSERT INTO tags (block_id, tag) VALUES (?, ?)", b.ID, tag); err != nil {
			return fmt.Errorf("inserting tag: %w", err)
		}
	}

	// Sync links
	if _, err := tx.Exec("DELETE FROM links WHERE source_id = ?", b.ID); err != nil {
		return fmt.Errorf("clearing links: %w", err)
	}
	seenLinks := make(map[string]bool)
	for _, link := range b.Links {
		// Try to resolve link (which might be a title or an ID) to a definitive ID
		var targetID string
		// 1. Is it already a valid ID?
		err := tx.QueryRow("SELECT id FROM blocks WHERE id = ?", link).Scan(&targetID)
		if err != nil {
			// 2. Is it a title?
			_ = tx.QueryRow("SELECT id FROM blocks WHERE title = ?", link).Scan(&targetID)
		}

		if targetID != "" && targetID != b.ID {
			if !seenLinks[targetID] {
				if _, err := tx.Exec("INSERT OR IGNORE INTO links (source_id, target_id) VALUES (?, ?)", b.ID, targetID); err != nil {
					return fmt.Errorf("inserting link %s -> %s: %w", b.ID, targetID, err)
				}
				seenLinks[targetID] = true
			}
		}
	}

	return tx.Commit()
}

// DeleteBlock removes a block from the index by ID.
func (s *Store) DeleteBlock(id string) error {
	_, err := s.db.Exec("DELETE FROM blocks WHERE id = ?", id)
	return err
}

// GetBlock retrieves a single block by ID.
func (s *Store) GetBlock(id string) (*block.Block, error) {
	row := s.db.QueryRow(`
		SELECT id, title, status, due, created, modified, filepath, body_preview, area
		FROM blocks WHERE id = ?
	`, id)
	return scanBlock(row)
}

// GetBlockByPath retrieves a block by its file path.
func (s *Store) GetBlockByPath(filepath string) (*block.Block, error) {
	row := s.db.QueryRow(`
		SELECT id, title, status, due, created, modified, filepath, body_preview, area
		FROM blocks WHERE filepath = ?
	`, filepath)
	return scanBlock(row)
}

// GetBlockByTitle retrieves the first block whose title matches exactly (case-insensitive).
// Used for transclusion of ![[title]] references.
func (s *Store) GetBlockByTitle(title string) (*block.Block, error) {
	row := s.db.QueryRow(`
		SELECT id, title, status, due, created, modified, filepath, body_preview, area
		FROM blocks WHERE LOWER(title) = LOWER(?) LIMIT 1
	`, title)
	return scanBlock(row)
}

// ListBlocks returns all blocks, optionally filtered.
func (s *Store) ListBlocks(filter Filter) ([]*block.Block, error) {
	query := "SELECT b.id, b.title, b.status, b.due, b.created, b.modified, b.filepath, b.body_preview, b.area FROM blocks b"
	var conditions []string
	var args []any

	if filter.Status != nil {
		conditions = append(conditions, "b.status = ?")
		args = append(args, *filter.Status)
	}
	if filter.HasStatus {
		conditions = append(conditions, "b.status != ''")
	}
	if filter.NoStatus {
		conditions = append(conditions, "b.status = ''")
	}
	if filter.Tag != "" {
		query += " JOIN tags t ON t.block_id = b.id"
		conditions = append(conditions, "t.tag = ?")
		args = append(args, filter.Tag)
	}
	if filter.DueBefore != nil {
		conditions = append(conditions, "b.due IS NOT NULL AND b.due <= ?")
		args = append(args, filter.DueBefore.Format(time.RFC3339))
	}
	if filter.DueAfter != nil {
		conditions = append(conditions, "b.due IS NOT NULL AND b.due >= ?")
		args = append(args, filter.DueAfter.Format(time.RFC3339))
	}
	if filter.Area != "" {
		conditions = append(conditions, "b.area = ?")
		args = append(args, filter.Area)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY b.modified DESC"
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing blocks: %w", err)
	}
	defer rows.Close()

	var blocks []*block.Block
	for rows.Next() {
		b, err := scanBlockRows(rows)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, b)
	}
	return blocks, rows.Err()
}

// Search performs full-text search on blocks using FTS5.
func (s *Store) Search(query string) ([]*block.Block, error) {
	// Format for FTS5 prefix match (e.g., "apple"* AND "pie"*)
	words := strings.Fields(query)
	for i, w := range words {
		words[i] = "\"" + strings.ReplaceAll(w, "\"", "\"\"") + "\"*"
	}
	ftsQuery := strings.Join(words, " AND ")
	if ftsQuery == "" {
		return nil, nil
	}

	rows, err := s.db.Query(`
		SELECT b.id, b.title, b.status, b.due, b.created, b.modified, b.filepath, 
			snippet(blocks_fts, -1, '»', '«', '...', 8) as body_preview, 
			b.area
		FROM blocks_fts f
		JOIN blocks b ON b.id = f.id
		WHERE blocks_fts MATCH ?
		ORDER BY rank
		LIMIT 50
	`, ftsQuery)
	if err != nil {
		return nil, fmt.Errorf("searching: %w", err)
	}
	defer rows.Close()

	var blocks []*block.Block
	for rows.Next() {
		b, err := scanBlockRows(rows)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, b)
	}
	return blocks, rows.Err()
}

// GetTags returns all unique tags in the index.
func (s *Store) GetTags() ([]string, error) {
	rows, err := s.db.Query("SELECT DISTINCT tag FROM tags ORDER BY tag")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

// GetTagCounts returns all unique tags and their block frequencies.
func (s *Store) GetTagCounts() (map[string]int, error) {
	rows, err := s.db.Query("SELECT tag, count(*) FROM tags GROUP BY tag ORDER BY count(*) DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var tag string
		var count int
		if err := rows.Scan(&tag, &count); err != nil {
			return nil, err
		}
		counts[tag] = count
	}
	return counts, rows.Err()
}

// GetLinksFrom returns IDs of blocks linked from the given block.
func (s *Store) GetLinksFrom(blockID string) ([]string, error) {
	rows, err := s.db.Query("SELECT target_id FROM links WHERE source_id = ?", blockID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// GetLinksTo returns IDs of blocks that link to the given block (backlinks).
func (s *Store) GetLinksTo(blockID string) ([]string, error) {
	rows, err := s.db.Query("SELECT source_id FROM links WHERE target_id = ?", blockID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// GraphEdge represents a connection between two blocks.
type GraphEdge struct {
	SourceID string
	TargetID string
}

// GetGraphEdges returns all documented links in the graph.
func (s *Store) GetGraphEdges() ([]GraphEdge, error) {
	rows, err := s.db.Query("SELECT source_id, target_id FROM links")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []GraphEdge
	for rows.Next() {
		var e GraphEdge
		if err := rows.Scan(&e.SourceID, &e.TargetID); err != nil {
			return nil, err
		}
		edges = append(edges, e)
	}
	return edges, rows.Err()
}

// FindRelatedBlocks uses FTS5 BM25 scoring to find blocks conceptually similar
// to the target block, explicitly excluding the target block itself.
func (s *Store) FindRelatedBlocks(target *block.Block, limit int) ([]*block.Block, error) {
	// Extract basic keywords (tags + words from title/body)
	var words []string
	words = append(words, target.Tags...)

	// Super naive tokenizer: split by space, keep words > 4 chars
	text := target.Title + " " + target.Body
	parts := strings.Fields(text)
	for _, p := range parts {
		// remove basic punctuation
		p = strings.Trim(p, ".,!?:;()[]{}\"'")
		if len(p) > 4 {
			// Double quote to prevent FTS syntax errors
			words = append(words, "\""+p+"\"")
		}
	}

	if len(words) == 0 {
		return nil, nil // Nothing to match against
	}

	// Limit term count to avoid massive queries
	if len(words) > 30 {
		words = words[:30]
	}

	matchQuery := strings.Join(words, " OR ")

	rows, err := s.db.Query(`
		SELECT b.id, b.title, b.status, b.due, b.created, b.modified, b.filepath, b.body_preview, b.area
		FROM blocks_fts f
		JOIN blocks b ON b.id = f.id
		WHERE blocks_fts MATCH ? AND b.id != ?
		ORDER BY rank
		LIMIT ?
	`, matchQuery, target.ID, limit)
	if err != nil {
		return nil, fmt.Errorf("finding related: %w", err)
	}
	defer rows.Close()

	var blocks []*block.Block
	for rows.Next() {
		b, err := scanBlockRows(rows)
		if err != nil {
			return nil, err
		}
		// FTS tables don't load tags/links automatically, so we might need to lazy load them
		blocks = append(blocks, b)
	}
	return blocks, rows.Err()
}

// BlockCount returns the total number of indexed blocks.
func (s *Store) BlockCount() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM blocks").Scan(&count)
	return count, err
}

// Filter configures block listing queries.
type Filter struct {
	Status    *string
	HasStatus bool // status != ""
	NoStatus  bool // status == ""
	Tag       string
	DueBefore *time.Time
	DueAfter  *time.Time
	Limit     int
	Area      string
}

// scanBlock scans a single block from a *sql.Row.
func scanBlock(row *sql.Row) (*block.Block, error) {
	var b block.Block
	var dueStr, createdStr, modifiedStr, bodyPreview, area sql.NullString

	err := row.Scan(&b.ID, &b.Title, &b.Status, &dueStr, &createdStr, &modifiedStr, &b.FilePath, &bodyPreview, &area)
	if err != nil {
		return nil, err
	}

	if dueStr.Valid {
		t, _ := time.Parse(time.RFC3339, dueStr.String)
		b.Due = &t
	}
	if createdStr.Valid {
		b.Created, _ = time.Parse(time.RFC3339, createdStr.String)
	}
	if modifiedStr.Valid {
		b.Modified, _ = time.Parse(time.RFC3339, modifiedStr.String)
	}
	if bodyPreview.Valid {
		b.Body = bodyPreview.String
	}
	if area.Valid {
		b.Area = area.String
	}
	return &b, nil
}

// scanBlockRows scans a block from *sql.Rows (same fields as scanBlock).
func scanBlockRows(rows *sql.Rows) (*block.Block, error) {
	var b block.Block
	var dueStr, createdStr, modifiedStr, bodyPreview, area sql.NullString

	err := rows.Scan(&b.ID, &b.Title, &b.Status, &dueStr, &createdStr, &modifiedStr, &b.FilePath, &bodyPreview, &area)
	if err != nil {
		return nil, err
	}

	if dueStr.Valid {
		t, _ := time.Parse(time.RFC3339, dueStr.String)
		b.Due = &t
	}
	if createdStr.Valid {
		b.Created, _ = time.Parse(time.RFC3339, createdStr.String)
	}
	if modifiedStr.Valid {
		b.Modified, _ = time.Parse(time.RFC3339, modifiedStr.String)
	}
	if bodyPreview.Valid {
		b.Body = bodyPreview.String
	}
	if area.Valid {
		b.Area = area.String
	}
	return &b, nil
}
