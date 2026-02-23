package main

import (
	"fmt"

	"github.com/charmbracelet/glamour"
	"github.com/lemondesk/neocognito/internal/tui/styles"
)

func main() {
	content := "### This is an H3\n\nSome text with *italics* and **bold**.\n\n# H1 Title\n\n## H2 Title"

	renderer, err := glamour.NewTermRenderer(
		glamour.WithStyles(styles.MarkdownStyle()),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		panic(err)
	}

	out, err := renderer.Render(content)
	if err != nil {
		panic(err)
	}

	fmt.Println("RENDERED OUTPUT:")
	fmt.Println(out)
}
