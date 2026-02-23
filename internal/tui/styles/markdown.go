package styles

import (
	"github.com/charmbracelet/glamour/ansi"
	gstyles "github.com/charmbracelet/glamour/styles"
)

// MarkdownStyle returns a custom Glamour style.
func MarkdownStyle() ansi.StyleConfig {
	style := gstyles.DarkStyleConfig

	// Remove markdown hashes from headings and make them look like titles
	style.H1.Prefix = " "
	style.H1.Suffix = " "

	style.H2.Prefix = " ▌ "
	style.H2.Suffix = " "
	style.H2.Color = stringPtr("212") // pinkish

	style.H3.Prefix = " ┃ "
	style.H3.Suffix = " "
	style.H3.Color = stringPtr("141") // purpleish

	style.H4.Prefix = " │ "
	style.H4.Suffix = " "

	style.H5.Prefix = " ┆ "
	style.H5.Suffix = " "

	style.H6.Prefix = " ┊ "
	style.H6.Suffix = " "

	style.Item.BlockPrefix = "• "

	return style
}

func stringPtr(s string) *string {
	return &s
}
