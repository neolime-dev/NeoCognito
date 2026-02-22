// Package styles provides the Lipgloss theme for the NeoCognito TUI.
package styles

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type ThemeColors struct {
	Primary   lipgloss.TerminalColor
	Secondary lipgloss.TerminalColor
	Accent    lipgloss.TerminalColor
	Success   lipgloss.TerminalColor
	Warning   lipgloss.TerminalColor
	Muted     lipgloss.TerminalColor
	Surface   lipgloss.TerminalColor
	Text      lipgloss.TerminalColor
	TextDim   lipgloss.TerminalColor
}

var (
	TokyoNight = ThemeColors{
		Primary:   lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7AA2F7"},
		Secondary: lipgloss.AdaptiveColor{Light: "#2B2A4C", Dark: "#E0DEF4"},
		Accent:    lipgloss.AdaptiveColor{Light: "#E95678", Dark: "#F7768E"},
		Success:   lipgloss.AdaptiveColor{Light: "#2D9574", Dark: "#9ECE6A"},
		Warning:   lipgloss.AdaptiveColor{Light: "#E0AF68", Dark: "#E0AF68"},
		Muted:     lipgloss.AdaptiveColor{Light: "#888888", Dark: "#565F89"},
		Surface:   lipgloss.AdaptiveColor{Light: "#F5F5F5", Dark: "#1A1B26"},
		Text:      lipgloss.AdaptiveColor{Light: "#2B2A4C", Dark: "#C0CAF5"},
		TextDim:   lipgloss.AdaptiveColor{Light: "#999999", Dark: "#414868"},
	}

	Catppuccin = ThemeColors{
		Primary:   lipgloss.AdaptiveColor{Light: "#1E66F5", Dark: "#89B4FA"},
		Secondary: lipgloss.AdaptiveColor{Light: "#E6E9EF", Dark: "#1E1E2E"},
		Accent:    lipgloss.AdaptiveColor{Light: "#D20F39", Dark: "#F38BA8"},
		Success:   lipgloss.AdaptiveColor{Light: "#40A02B", Dark: "#A6E3A1"},
		Warning:   lipgloss.AdaptiveColor{Light: "#DF8E1D", Dark: "#F9E2AF"},
		Muted:     lipgloss.AdaptiveColor{Light: "#9CA0B0", Dark: "#6C7086"},
		Surface:   lipgloss.AdaptiveColor{Light: "#EFF1F5", Dark: "#11111B"},
		Text:      lipgloss.AdaptiveColor{Light: "#4C4F69", Dark: "#CDD6F4"},
		TextDim:   lipgloss.AdaptiveColor{Light: "#7C7F93", Dark: "#A6ADC8"},
	}

	Nord = ThemeColors{
		Primary:   lipgloss.AdaptiveColor{Light: "#5E81AC", Dark: "#81A1C1"},
		Secondary: lipgloss.AdaptiveColor{Light: "#ECEFF4", Dark: "#3B4252"},
		Accent:    lipgloss.AdaptiveColor{Light: "#BF616A", Dark: "#BF616A"},
		Success:   lipgloss.AdaptiveColor{Light: "#A3BE8C", Dark: "#A3BE8C"},
		Warning:   lipgloss.AdaptiveColor{Light: "#EBCB8B", Dark: "#EBCB8B"},
		Muted:     lipgloss.AdaptiveColor{Light: "#D8DEE9", Dark: "#4C566A"},
		Surface:   lipgloss.AdaptiveColor{Light: "#E5E9F0", Dark: "#2E3440"},
		Text:      lipgloss.AdaptiveColor{Light: "#2E3440", Dark: "#ECEFF4"},
		TextDim:   lipgloss.AdaptiveColor{Light: "#4C566A", Dark: "#D8DEE9"},
	}

	Gruvbox = ThemeColors{
		Primary:   lipgloss.AdaptiveColor{Light: "#458588", Dark: "#83A598"},
		Secondary: lipgloss.AdaptiveColor{Light: "#EBDBB2", Dark: "#3C3836"},
		Accent:    lipgloss.AdaptiveColor{Light: "#CC241D", Dark: "#FB4934"},
		Success:   lipgloss.AdaptiveColor{Light: "#98971A", Dark: "#B8BB26"},
		Warning:   lipgloss.AdaptiveColor{Light: "#D79921", Dark: "#FABD2F"},
		Muted:     lipgloss.AdaptiveColor{Light: "#928374", Dark: "#928374"},
		Surface:   lipgloss.AdaptiveColor{Light: "#FBF1C7", Dark: "#282828"},
		Text:      lipgloss.AdaptiveColor{Light: "#3C3836", Dark: "#EBDBB2"},
		TextDim:   lipgloss.AdaptiveColor{Light: "#7C6F64", Dark: "#A89984"},
	}
)

var (
	Primary   lipgloss.TerminalColor
	Secondary lipgloss.TerminalColor
	Accent    lipgloss.TerminalColor
	Success   lipgloss.TerminalColor
	Warning   lipgloss.TerminalColor
	Muted     lipgloss.TerminalColor
	Surface   lipgloss.TerminalColor
	Text      lipgloss.TerminalColor
	TextDim   lipgloss.TerminalColor
)

var (
	ActiveBorder      lipgloss.Style
	InactiveBorder    lipgloss.Style
	TitleStyle        lipgloss.Style
	PrimaryStyle      lipgloss.Style
	SuccessStyle      lipgloss.Style
	AccentStyle       lipgloss.Style
	StatusBarStyle    lipgloss.Style
	KeyStyle          lipgloss.Style
	DescStyle         lipgloss.Style
	NormalModeStyle   lipgloss.Style
	InsertModeStyle   lipgloss.Style
	SelectedItemStyle lipgloss.Style
	NormalItemStyle   lipgloss.Style
	DimItemStyle      lipgloss.Style
	TodoBadge         lipgloss.Style
	DoingBadge        lipgloss.Style
	DoneBadge         lipgloss.Style
	ArchivedBadge     lipgloss.Style
	TagStyle          lipgloss.Style
	SearchPromptStyle lipgloss.Style
	SearchMatchStyle  lipgloss.Style
)

func LoadTheme(name string) {
	colors := TokyoNight
	switch name {
	case "catppuccin":
		colors = Catppuccin
	case "nord":
		colors = Nord
	case "gruvbox":
		colors = Gruvbox
	case "omarchy", "system":
		colors = loadOmarchy()
	case "tokyo-night", "default", "":
		colors = TokyoNight
	}

	Primary = colors.Primary
	Secondary = colors.Secondary
	Accent = colors.Accent
	Success = colors.Success
	Warning = colors.Warning
	Muted = colors.Muted
	Surface = colors.Surface
	Text = colors.Text
	TextDim = colors.TextDim

	ActiveBorder = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(Primary)
	InactiveBorder = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(Muted)
	TitleStyle = lipgloss.NewStyle().Foreground(Primary).Bold(true).Padding(0, 1)
	PrimaryStyle = lipgloss.NewStyle().Foreground(Primary)
	SuccessStyle = lipgloss.NewStyle().Foreground(Success)
	AccentStyle = lipgloss.NewStyle().Foreground(Accent)
	StatusBarStyle = lipgloss.NewStyle().Foreground(TextDim).Padding(0, 1)
	KeyStyle = lipgloss.NewStyle().Foreground(Primary).Bold(true)
	DescStyle = lipgloss.NewStyle().Foreground(Muted)
	NormalModeStyle = lipgloss.NewStyle().Background(Primary).Foreground(lipgloss.Color("#FFFFFF")).Padding(0, 1).Bold(true)
	InsertModeStyle = lipgloss.NewStyle().Background(Success).Foreground(lipgloss.Color("#FFFFFF")).Padding(0, 1).Bold(true)
	SelectedItemStyle = lipgloss.NewStyle().Foreground(Primary).Bold(true)
	NormalItemStyle = lipgloss.NewStyle().Foreground(Text)
	DimItemStyle = lipgloss.NewStyle().Foreground(Muted)

	TodoBadge = lipgloss.NewStyle().Foreground(Warning).Bold(true).SetString("○ TODO")
	DoingBadge = lipgloss.NewStyle().Foreground(Primary).Bold(true).SetString("◔ DOING")
	DoneBadge = lipgloss.NewStyle().Foreground(Success).Bold(true).SetString("● DONE")
	ArchivedBadge = lipgloss.NewStyle().Foreground(Muted).SetString("◌ ARCHIVED")

	TagStyle = lipgloss.NewStyle().Foreground(Accent).Italic(true)
	SearchPromptStyle = lipgloss.NewStyle().Foreground(Primary).Bold(true)
	SearchMatchStyle = lipgloss.NewStyle().Foreground(Accent).Bold(true)
}

func init() {
	// Initialize with default theme so tests and non-configured runs don't panic
	LoadTheme("tokyo-night")
}

// loadOmarchy attempts to read the active Omarchy OS theme via ~/.config/omarchy/current/theme/kitty.conf
// and map its 16 ANSI colors into the NeoCognito semantic Palette.
func loadOmarchy() ThemeColors {
	// Fallback to TokyoNight if something goes wrong
	fallback := TokyoNight

	home, err := os.UserHomeDir()
	if err != nil {
		return fallback
	}

	confPath := filepath.Join(home, ".config", "omarchy", "current", "theme", "kitty.conf")
	file, err := os.Open(confPath)
	if err != nil {
		return fallback
	}
	defer file.Close()

	// Parse out hex codes
	hexes := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			hexes[parts[0]] = parts[1]
		}
	}

	// Helper to extract a color safely
	getColor := func(key string, fallbackColor lipgloss.TerminalColor) lipgloss.TerminalColor {
		if hex, ok := hexes[key]; ok {
			return lipgloss.Color(hex)
		}
		return fallbackColor
	}

	return ThemeColors{
		Primary:   getColor("color4", fallback.Primary),     // Blue
		Secondary: getColor("color5", fallback.Secondary),   // Magenta
		Accent:    getColor("color6", fallback.Accent),      // Cyan
		Success:   getColor("color2", fallback.Success),     // Green
		Warning:   getColor("color3", fallback.Warning),     // Yellow
		Muted:     getColor("color8", fallback.Muted),       // Bright Black/Gray
		Surface:   getColor("background", fallback.Surface), // BG
		Text:      getColor("foreground", fallback.Text),    // FG
		TextDim:   getColor("color7", fallback.TextDim),     // White/Dim
	}
}

// StatusBadge returns the styled badge for a given status string.
func StatusBadge(status string) string {
	switch status {
	case "todo":
		return TodoBadge.String()
	case "doing":
		return DoingBadge.String()
	case "done":
		return DoneBadge.String()
	case "archived":
		return ArchivedBadge.String()
	default:
		return ""
	}
}
