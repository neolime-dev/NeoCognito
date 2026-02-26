// Package graphvis provides a horizontal Mind-Map / GitGraph visualization.
package graphvis

import (
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/neolime-dev/neocognito/internal/block"
	"github.com/neolime-dev/neocognito/internal/store"
	"github.com/neolime-dev/neocognito/internal/tui/styles"
)

// Msg and cmds
type OpenGraphMsg struct{}
type CloseGraphMsg struct{}

// Node represents a block in the mind-map tree
type Node struct {
	Block *block.Block
	X, Y  float64
	Level int
}

type Model struct {
	Active      bool
	Width       int
	Height      int
	forceRun    bool
	initialized bool

	nodes   []*Node
	edges   []store.GraphEdge
	nodeMap map[string]*Node

	cameraX, cameraY float64
	zoom             float64
	cursor           int // index of selected node

	// Callbacks
	GetBlocksFn func() ([]*block.Block, error)
	GetEdgesFn  func() ([]store.GraphEdge, error)
	OnSelect    func(*block.Block) tea.Cmd

	// State
	needsCentering bool
	FocusedID      string
	listOffset     int
}

func New() Model {
	return Model{nodeMap: make(map[string]*Node), forceRun: true, zoom: 1.0}
}

func (m *Model) LoadData() {
	if m.GetBlocksFn == nil || m.GetEdgesFn == nil {
		return
	}

	blocks, _ := m.GetBlocksFn()
	m.edges, _ = m.GetEdgesFn()

	m.nodes = nil
	m.nodeMap = make(map[string]*Node)

	// Re-initialize nodes
	for _, b := range blocks {
		n := &Node{Block: b}
		m.nodes = append(m.nodes, n)
		m.nodeMap[b.ID] = n
	}

	// Initialize cursor to the most connected node on first load
	if !m.initialized && len(m.nodes) > 0 {
		degrees := make(map[string]int)
		for _, e := range m.edges {
			degrees[e.SourceID]++
			degrees[e.TargetID]++
		}
		maxDeg := -1
		bestIdx := 0
		for i, n := range m.nodes {
			if degrees[n.Block.ID] > maxDeg {
				maxDeg = degrees[n.Block.ID]
				bestIdx = i
			}
		}
		m.cursor = bestIdx
		if bestIdx < len(m.nodes) {
			m.FocusedID = m.nodes[bestIdx].Block.ID
		}
		m.initialized = true
	}

	// Always ensure cursor is valid
	if len(m.nodes) > 0 {
		if m.FocusedID != "" {
			for i, n := range m.nodes {
				if n.Block.ID == m.FocusedID {
					m.cursor = i
					break
				}
			}
		}

		if m.cursor >= len(m.nodes) {
			m.cursor = 0
		}
		m.FocusedID = m.nodes[m.cursor].Block.ID
		m.runMindMapLayout()
	}
	m.forceRun = false
}

// runMindMapLayout organizes nodes in a left-to-right hierarchy centered on the focused node
func (m *Model) runMindMapLayout() {
	if len(m.nodes) == 0 {
		return
	}

	focusedID := m.nodes[m.cursor].Block.ID
	levelMap := make(map[string]int)
	for _, n := range m.nodes {
		levelMap[n.Block.ID] = -999 // Unvisited
	}
	levelMap[focusedID] = 0

	// BFS for forward links (consequences -> right)
	queue := []string{focusedID}
	visited := make(map[string]bool)
	visited[focusedID] = true
	for len(queue) > 0 {
		currID := queue[0]
		queue = queue[1:]
		currLevel := levelMap[currID]

		for _, e := range m.edges {
			if e.SourceID == currID {
				if levelMap[e.TargetID] == -999 {
					levelMap[e.TargetID] = currLevel + 1
					queue = append(queue, e.TargetID)
				}
			}
		}
	}

	// BFS for backlinks (influences -> left)
	queue = []string{focusedID}
	visitedBack := make(map[string]bool)
	visitedBack[focusedID] = true
	for len(queue) > 0 {
		currID := queue[0]
		queue = queue[1:]
		currLevel := levelMap[currID]

		for _, e := range m.edges {
			if e.TargetID == currID {
				if levelMap[e.SourceID] == -999 {
					levelMap[e.SourceID] = currLevel - 1
					queue = append(queue, e.SourceID)
				}
			}
		}
	}

	// Group by level and assign coordinates
	levelGroups := make(map[int][]*Node)
	for _, n := range m.nodes {
		l := levelMap[n.Block.ID]
		if l == -999 {
			l = 5 // Put disconnected nodes to the far right
		}
		n.Level = l
		levelGroups[l] = append(levelGroups[l], n)
	}

	colWidth := 50.0
	rowHeight := 6.0

	for l, nodes := range levelGroups {
		for i, n := range nodes {
			n.X = float64(l) * colWidth
			n.Y = float64(i-len(nodes)/2) * rowHeight
		}
	}

	// Center camera on focused node ONLY if requested
	if m.needsCentering {
		m.cameraX = m.nodes[m.cursor].X
		m.cameraY = m.nodes[m.cursor].Y
		m.needsCentering = false
	}
}

// runLayout is no longer used for the mind-map pivot
func (m *Model) runLayout(iterations int) {}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if m.forceRun {
		m.LoadData()
	}

	if !m.Active {
		return *m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if len(m.nodes) > 0 {
				m.cursor--
				if m.cursor < 0 {
					m.cursor = len(m.nodes) - 1
				}
				m.FocusedID = m.nodes[m.cursor].Block.ID
				m.needsCentering = true
				m.runMindMapLayout()
			}
		case "down", "j":
			if len(m.nodes) > 0 {
				m.cursor = (m.cursor + 1) % len(m.nodes)
				m.FocusedID = m.nodes[m.cursor].Block.ID
				m.needsCentering = true
				m.runMindMapLayout()
			}
		case "left", "h":
			m.cameraX -= 20
		case "right", "l":
			m.cameraX += 20
		case "enter":
			if len(m.nodes) > 0 && m.cursor < len(m.nodes) && m.OnSelect != nil {
				return *m, m.OnSelect(m.nodes[m.cursor].Block)
			}
		case "+", "=":
			m.zoom *= 1.5
		case "-":
			m.zoom /= 1.5
		}
	}
	return *m, nil
}

func (m Model) View() string {
	if m.Width < 10 || m.Height < 10 {
		return ""
	}

	listWidth := 30
	if m.Width < 50 {
		listWidth = m.Width / 3 // fallback for very small screens
	}
	graphWidth := m.Width - listWidth - 4
	graphHeight := m.Height - 6

	// 1. Build Left Pane (Node List)
	var listSb strings.Builder
	listSb.WriteString(styles.TitleStyle.Render("󰈚 Nodes") + "\n\n")

	// Adjust list offset for scrolling
	visibleItems := m.Height - 8
	if visibleItems < 1 {
		visibleItems = 1
	}
	if m.cursor < m.listOffset {
		m.listOffset = m.cursor
	} else if m.cursor >= m.listOffset+visibleItems {
		m.listOffset = m.cursor - visibleItems + 1
	}

	for i := m.listOffset; i < m.listOffset+visibleItems && i < len(m.nodes); i++ {
		n := m.nodes[i]
		prefix := "  "
		title := n.Block.Title
		if len(title) > listWidth-6 {
			title = title[:listWidth-8] + ".."
		}

		if i == m.cursor {
			prefix = "► "
			listSb.WriteString(styles.SuccessStyle.Render(prefix+title) + "\n")
		} else {
			listSb.WriteString(styles.NormalModeStyle.Render(prefix+title) + "\n")
		}
	}
	leftPane := lipgloss.NewStyle().
		Width(listWidth).
		Height(m.Height-4).
		Border(lipgloss.NormalBorder(), false, true, false, false).
		BorderForeground(styles.Muted).
		PaddingRight(1).
		Render(listSb.String())

	// 2. Build Right Pane (Graph)
	// Create an empty screen buffer for the graph half
	screenW, screenH := graphWidth, graphHeight
	buffer := make([][]rune, screenH)
	for i := range buffer {
		buffer[i] = make([]rune, screenW)
		for j := range buffer[i] {
			buffer[i][j] = ' '
		}
	}

	scale := 1.5 * m.zoom // Scale with zoom factor

	// Helper to draw a character
	drawPoint := func(x, y int, char rune) {
		if x >= 0 && x < screenW && y >= 0 && y < screenH {
			buffer[y][x] = char
		}
	}

	// Bresenham's line drawing algorithm for ASCII
	drawLine := func(x0, y0, x1, y1 int, char rune) {
		dx := math.Abs(float64(x1 - x0))
		dy := math.Abs(float64(y1 - y0))
		sx, sy := -1, -1
		if x0 < x1 {
			sx = 1
		}
		if y0 < y1 {
			sy = 1
		}
		err := dx - dy

		for {
			drawPoint(x0, y0, char)
			if x0 == x1 && y0 == y1 {
				break
			}
			e2 := 2 * err
			if e2 > -dy {
				err -= dy
				x0 += sx
			}
			if e2 < dx {
				err += dx
				y0 += sy
			}
		}
	}

	// Draw edges (Mind-Map style branching)
	for _, e := range m.edges {
		n1, ok1 := m.nodeMap[e.SourceID]
		n2, ok2 := m.nodeMap[e.TargetID]
		if !ok1 || !ok2 {
			continue
		}
		sx := int((n1.X-m.cameraX)*scale) + screenW/2
		sy := int((n1.Y-m.cameraY)*scale) + screenH/2
		tx := int((n2.X-m.cameraX)*scale) + screenW/2
		ty := int((n2.Y-m.cameraY)*scale) + screenH/2

		char := '·'
		isHighlighted := false
		if len(m.nodes) > 0 {
			focusedID := m.nodes[m.cursor].Block.ID
			if e.SourceID == focusedID || e.TargetID == focusedID {
				isHighlighted = true
				char = '─'
			}
		}

		// Draw a horizontal branch if nodes are on different levels
		if sx != tx {
			midX := sx + (tx-sx)/2
			drawLine(sx, sy, midX, sy, char)
			drawLine(midX, sy, midX, ty, char)
			drawLine(midX, ty, tx, ty, char)

			// Add "elbow" joints
			if isHighlighted {
				if sy < ty {
					drawPoint(midX, sy, '╮')
					drawPoint(midX, ty, '╰')
				} else if sy > ty {
					drawPoint(midX, sy, '╯')
					drawPoint(midX, ty, '╭')
				}
			}
		} else {
			drawLine(sx, sy, tx, ty, char)
		}
	}

	// Draw nodes (Layer 1: Markers)
	for i, n := range m.nodes {
		nx := int((n.X-m.cameraX)*scale) + screenW/2
		ny := int((n.Y-m.cameraY)*scale) + screenH/2

		marker := '●'
		if i == m.cursor {
			marker = '⦿'
		} else if len(m.nodes) > 0 {
			// Highlight neighbors
			focusedID := m.nodes[m.cursor].Block.ID
			isNeighbor := false
			for _, e := range m.edges {
				if (e.SourceID == focusedID && e.TargetID == n.Block.ID) ||
					(e.TargetID == focusedID && e.SourceID == n.Block.ID) {
					isNeighbor = true
					break
				}
			}
			if isNeighbor {
				marker = '•'
			}
		}
		drawPoint(nx, ny, marker)
	}

	// Draw labels (Layer 2: Text)
	for i, n := range m.nodes {
		nx := int((n.X-m.cameraX)*scale) + screenW/2
		ny := int((n.Y-m.cameraY)*scale) + screenH/2

		title := n.Block.Title
		if i == m.cursor {
			// Selected node: Add a clear prefix
			fullTitle := "► " + title
			for j, char := range fullTitle {
				drawPoint(nx+2+j, ny, char)
			}
		} else {
			// Other nodes: Truncated dynamically based on zoom
			limit := int(15.0 * m.zoom)
			if limit < 5 {
				limit = 5 // Minimum limit to avoid negative slicing
			}
			if len(title) > limit {
				title = title[:limit-2] + ".."
			}
			for j, char := range title {
				drawPoint(nx+2+j, ny, char)
			}
		}
	}

	// Convert buffer to string
	var sb strings.Builder
	for y := 0; y < screenH; y++ {
		sb.WriteString(string(buffer[y]) + "\n")
	}

	renderedGrid := styles.PrimaryStyle.Render(strings.TrimRight(sb.String(), "\n"))

	header := styles.TitleStyle.Render("󱎹 Knowledge Map (Mind-Map Mode)")

	focusedTitle := ""
	if len(m.nodes) > 0 && m.cursor < len(m.nodes) {
		focusedTitle = styles.SuccessStyle.Render("  FOCUS: " + m.nodes[m.cursor].Block.Title)
	}

	help := styles.DimItemStyle.Render("  [j/k] list nav  [h/l] manual pan  [+/-] zoom  [enter] open block")

	rightPane := lipgloss.JoinVertical(lipgloss.Left, header+"\n", focusedTitle, renderedGrid, "\n"+help)

	content := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

	return styles.ActiveBorder.
		Width(m.Width).
		Height(m.Height).
		Render(content)
}

func (m *Model) SetSize(w, h int) {
	m.Width = w
	m.Height = h
}
