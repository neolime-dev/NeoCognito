// Package export provides static HTML site generation for NeoCognito.
package export

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/neolime-dev/neocognito/internal/block"
	"github.com/neolime-dev/neocognito/internal/store"
	"github.com/neolime-dev/neocognito/internal/tui/styles"
	"github.com/yuin/goldmark"
)
// Note: fmt is kept for error formatting in Run.

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>{{.Title}} - NeoCognito</title>
    <style>
        :root {
            --primary: {{.Theme.Primary}};
            --secondary: {{.Theme.Secondary}};
            --accent: {{.Theme.Accent}};
            --success: {{.Theme.Success}};
            --warning: {{.Theme.Warning}};
            --muted: {{.Theme.Muted}};
            --surface: {{.Theme.Surface}};
            --text: {{.Theme.Text}};
            --text-dim: {{.Theme.TextDim}};
        }
        body {
            background-color: var(--surface);
            color: var(--text);
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
            line-height: 1.6;
            max-width: 800px;
            margin: 0 auto;
            padding: 2rem;
        }
        a { color: var(--primary); text-decoration: none; }
        a:hover { text-decoration: underline; }
        .header { border-bottom: 2px solid var(--primary); padding-bottom: 1rem; margin-bottom: 2rem; }
        .tag { background: var(--muted); color: var(--surface); padding: 0.2rem 0.5rem; border-radius: 4px; font-size: 0.8rem; margin-right: 0.5rem; }
        .meta { color: var(--text-dim); font-size: 0.9rem; margin-bottom: 1rem; }
        pre { background: #1a1b26; color: #a9b1d6; padding: 1rem; border-radius: 8px; overflow-x: auto; }
        code { font-family: "JetBrains Mono", monospace; }
        .nav { margin-bottom: 2rem; }
    </style>
</head>
<body>
    <div class="nav"><a href="index.html">← Back to Index</a></div>
    <div class="header">
        <h1>{{.Title}}</h1>
        <div class="meta">
            Created: {{.Created}} | Modified: {{.Modified}}
            {{if .Area}}<br>Area: {{.Area}}{{end}}
        </div>
        <div>
            {{range .Tags}}<span class="tag">#{{.}}</span>{{end}}
        </div>
    </div>
    <div class="content">
        {{.ContentHTML}}
    </div>
</body>
</html>
`

const indexTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>My Second Brain - NeoCognito</title>
    <style>
        :root {
            --primary: {{.Theme.Primary}};
            --surface: {{.Theme.Surface}};
            --text: {{.Theme.Text}};
            --text-dim: {{.Theme.TextDim}};
        }
        body { background: var(--surface); color: var(--text); font-family: sans-serif; max-width: 800px; margin: 0 auto; padding: 2rem; }
        h1 { border-bottom: 2px solid var(--primary); }
        .block-list { list-style: none; padding: 0; }
        .block-item { margin-bottom: 1rem; border-left: 4px solid var(--primary); padding-left: 1rem; }
        .block-title { font-weight: bold; font-size: 1.2rem; }
        .block-meta { font-size: 0.8rem; color: var(--text-dim); }
        a { color: var(--primary); text-decoration: none; }
    </style>
</head>
<body>
    <h1>🧠 My Knowledge Graph</h1>
    <ul class="block-list">
        {{range .Blocks}}
        <li class="block-item">
            <a href="{{.ID}}.html" class="block-title">{{.Title}}</a>
            <div class="block-meta">{{.Status}} | {{.Created}}</div>
        </li>
        {{end}}
    </ul>
</body>
</html>
`

type themeCSS struct {
	Primary, Secondary, Accent, Success, Warning, Muted, Surface, Text, TextDim template.CSS
}

type exportData struct {
	Title       string
	Created     string
	Modified    string
	Area        string
	Tags        []string
	ContentHTML template.HTML
	Theme       themeCSS
}

// Run executes the full site export.
func Run(st store.Storer, outDir string) error {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}

	blocks, err := st.ListBlocks(store.Filter{})
	if err != nil {
		return err
	}

	css := styles.CurrentTheme().CSSTheme()
	themeData := themeCSS{
		Primary:   template.CSS(css.Primary),
		Secondary: template.CSS(css.Secondary),
		Accent:    template.CSS(css.Accent),
		Success:   template.CSS(css.Success),
		Warning:   template.CSS(css.Warning),
		Muted:     template.CSS(css.Muted),
		Surface:   template.CSS(css.Surface),
		Text:      template.CSS(css.Text),
		TextDim:   template.CSS(css.TextDim),
	}

	tmpl, err := template.New("block").Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("parsing block template: %w", err)
	}
	md := goldmark.New()

	for _, b := range blocks {
		// soft-fail: skip blocks whose source file cannot be read
		full, err := block.ParseFile(b.FilePath)
		if err != nil {
			continue
		}

		var buf bytes.Buffer
		if err := md.Convert([]byte(full.Body), &buf); err != nil {
			continue // soft-fail: skip blocks whose body cannot be rendered
		}

		data := exportData{
			Title:       full.Title,
			Created:     full.Created.Format("2006-01-02"),
			Modified:    full.Modified.Format("2006-01-02"),
			Area:        full.Area,
			Tags:        full.Tags,
			ContentHTML: template.HTML(buf.String()),
			Theme:       themeData,
		}

		f, err := os.Create(filepath.Join(outDir, b.ID+".html"))
		if err != nil {
			return fmt.Errorf("creating %s.html: %w", b.ID, err)
		}
		if err := tmpl.Execute(f, data); err != nil {
			f.Close()
			return fmt.Errorf("executing template for %s: %w", b.ID, err)
		}
		f.Close()
	}

	// Index page
	indexTmpl, err := template.New("index").Parse(indexTemplate)
	if err != nil {
		return fmt.Errorf("parsing index template: %w", err)
	}
	f, err := os.Create(filepath.Join(outDir, "index.html"))
	if err != nil {
		return fmt.Errorf("creating index.html: %w", err)
	}
	if err := indexTmpl.Execute(f, struct {
		Blocks []*block.Block
		Theme  themeCSS
	}{
		Blocks: blocks,
		Theme:  themeData,
	}); err != nil {
		f.Close()
		return fmt.Errorf("executing index template: %w", err)
	}
	f.Close()

	return nil
}
