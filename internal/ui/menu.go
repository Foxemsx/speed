package ui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	apptheme "github.com/Foxemsx/riptide/internal/theme"
)

// Layout thresholds for a responsive menu.
const (
	horizontalThreshold = 100 // below this → vertical stack
	gridThreshold       = 88  // wide enough for 2×2 grid
	menuTickInterval    = 100 * time.Millisecond
)

// screenID identifies which destination the menu routes to.
type screenID int

const (
	screenMenu screenID = iota
	screenTest
	screenMonitor
	screenSettings
	screenExit
)

// menuItem is one selectable box in the startup menu.
type menuItem struct {
	title    string
	subtitle string
	screen   screenID
	hotkey   string
	features []string
	badge    string
}

// menuModel is the startup screen.
type menuModel struct {
	theme   apptheme.Theme
	compact bool
	width   int
	height  int
	cursor  int
	hovered int
	pulse   float64
	spinner spinner.Model
	items   []menuItem
}

func newMenuModel(theme apptheme.Theme, compact bool) *menuModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.Highlight)
	return &menuModel{
		theme:   theme,
		compact: compact,
		cursor:  0,
		hovered: -1,
		spinner: s,
		items: []menuItem{
			{
				title: "Speed Test", subtitle: "one-shot DL · UL · ping", screen: screenTest, hotkey: "1",
				features: []string{"Download + upload + latency", "~10s timed phases", "Save & compare runs"},
				badge:    "ONE-SHOT",
			},
			{
				title: "Bandwidth", subtitle: "live monitor · real traffic", screen: screenMonitor, hotkey: "2",
				features: []string{"Real PC throughput", "Session peaks", "Zero generated traffic"},
				badge:    "LIVE",
			},
			{
				title: "Settings", subtitle: "themes · history · install", screen: screenSettings, hotkey: "3",
				features: []string{"11 color themes", "Searchable settings", "Database & uninstall"},
				badge:    "TUNE",
			},
			{
				title: "Exit", subtitle: "quit riptide cleanly", screen: screenExit, hotkey: "4",
				features: []string{"Cancel any running test", "Clean shutdown", "See you next wave"},
				badge:    "",
			},
		},
	}
}

func (m *menuModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.tickCmd())
}

func (m *menuModel) tickCmd() tea.Cmd {
	return tea.Tick(menuTickInterval, func(time.Time) tea.Msg { return menuTickMsg{} })
}

type menuTickMsg struct{}

func (m *menuModel) Update(msg tea.Msg) (tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return tea.Quit, false
		case "left", "h":
			m.move(-1)
			return nil, false
		case "right", "l":
			m.move(1)
			return nil, false
		case "up", "k":
			m.moveUp()
			return nil, false
		case "down", "j":
			m.moveDown()
			return nil, false
		case "1", "2", "3", "4":
			for i, it := range m.items {
				if it.hotkey == msg.String() {
					m.cursor = i
					return m.selectCurrent(), false
				}
			}
		case "enter", " ":
			return m.selectCurrent(), false
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return nil, false
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return cmd, false
	case menuTickMsg:
		m.pulse = m.pulse + 0.08
		if m.pulse > 1 {
			m.pulse = 0
		}
		return m.tickCmd(), false
	case tea.MouseMsg:
		switch {
		case msg.Action == tea.MouseActionMotion:
			hit := -1
			for i, box := range m.boxRects() {
				if msg.X >= box.x && msg.X < box.x+box.w &&
					msg.Y >= box.y && msg.Y < box.y+box.h {
					hit = i
					break
				}
			}
			if hit != m.hovered {
				m.hovered = hit
				return nil, false
			}
		case msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress:
			for i, box := range m.boxRects() {
				if msg.X >= box.x && msg.X < box.x+box.w &&
					msg.Y >= box.y && msg.Y < box.y+box.h {
					m.cursor = i
					m.hovered = -1
					return m.selectCurrent(), false
				}
			}
		}
	}
	return nil, false
}

func (m *menuModel) move(delta int) {
	m.cursor = (m.cursor + delta + len(m.items)) % len(m.items)
	m.hovered = -1
}

func (m *menuModel) moveUp() {
	mode, _, _, _, _, _ := m.computeLayout()
	if mode == "grid" {
		if m.cursor >= 2 {
			m.cursor -= 2
		} else {
			m.cursor += 2
			if m.cursor >= len(m.items) {
				m.cursor = len(m.items) - 1
			}
		}
	} else {
		m.move(-1)
	}
	m.hovered = -1
}

func (m *menuModel) moveDown() {
	mode, _, _, _, _, _ := m.computeLayout()
	if mode == "grid" {
		if m.cursor < 2 {
			m.cursor += 2
			if m.cursor >= len(m.items) {
				m.cursor = len(m.items) - 1
			}
		} else {
			m.cursor -= 2
		}
	} else {
		m.move(1)
	}
	m.hovered = -1
}

func (m *menuModel) selectCurrent() tea.Cmd {
	switch m.items[m.cursor].screen {
	case screenTest:
		return menuSelectCmd(screenTest)
	case screenMonitor:
		return menuSelectCmd(screenMonitor)
	case screenSettings:
		return menuSelectCmd(screenSettings)
	default:
		return tea.Quit
	}
}

type boxRect struct{ x, y, w, h int }

func (m *menuModel) headerHeight() int {
	if m.compact {
		return 4
	}
	return 10
}

func (m *menuModel) computeLayout() (mode string, boxW, boxH, startY, startX int, gap int) {
	w, h := m.width, m.height
	if w <= 0 {
		w = 100
	}
	if h <= 0 {
		h = 30
	}
	gap = 2
	boxH = 14 // room for badge + bottom pad so labels stay inside the fill
	num := len(m.items)

	if num == 4 && w >= gridThreshold {
		mode = "grid"
		boxW = min((w-8-gap)/2, 34)
		if boxW < 22 {
			boxW = 22
		}
		totalW := boxW*2 + gap
		totalH := m.headerHeight() + 1 + boxH*2 + gap + 2
		startY = (h - totalH) / 2
		if startY < 0 {
			startY = 0
		}
		startX = (w - totalW) / 2
		if startX < 0 {
			startX = 0
		}
		return
	}

	boxW = m.boxWidth(w, num)
	mode = "horizontal"
	if w < horizontalThreshold {
		mode = "vertical"
		boxW = min(w-6, 48)
	}

	totalW := num * boxW
	if mode != "vertical" {
		totalW += (num - 1) * gap
	}
	stackH := m.headerHeight() + 1
	if mode == "vertical" {
		stackH += num*boxH + (num - 1)
	} else {
		stackH += boxH
	}
	startY = (h - stackH) / 2
	if startY < 0 {
		startY = 0
	}
	startX = (w - totalW) / 2
	if startX < 0 {
		startX = 0
	}
	return
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (m *menuModel) boxRects() []boxRect {
	mode, boxW, boxH, startY, startX, gap := m.computeLayout()
	rects := make([]boxRect, len(m.items))
	boxesY := startY + m.headerHeight() + 1

	switch mode {
	case "vertical":
		for i := range m.items {
			rects[i] = boxRect{x: startX, y: boxesY + i*(boxH+1), w: boxW, h: boxH}
		}
	case "grid":
		for i := range m.items {
			col := i % 2
			row := i / 2
			rects[i] = boxRect{
				x: startX + col*(boxW+gap),
				y: boxesY + row*(boxH+gap),
				w: boxW,
				h: boxH,
			}
		}
	default:
		for i := range m.items {
			rects[i] = boxRect{x: startX + i*(boxW+gap), y: boxesY, w: boxW, h: boxH}
		}
	}
	return rects
}

func (m *menuModel) boxWidth(termW, num int) int {
	if num < 1 {
		num = 4
	}
	maxEach := 28
	each := (termW - 4 - (num-1)*2) / num
	if each > maxEach {
		each = maxEach
	}
	if each < 18 {
		each = 18
	}
	return each
}

func (m *menuModel) View() string {
	mode, boxW, _, _, _, gap := m.computeLayout()

	boxes := make([]string, len(m.items))
	for i, it := range m.items {
		boxes[i] = m.renderBox(i, it, boxW)
	}

	var cards string
	switch mode {
	case "vertical":
		parts := make([]string, len(boxes))
		for i, b := range boxes {
			if i < len(boxes)-1 {
				parts[i] = lipgloss.NewStyle().MarginBottom(1).Render(b)
			} else {
				parts[i] = b
			}
		}
		cards = lipgloss.JoinVertical(lipgloss.Left, parts...)
	case "grid":
		row0 := lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().MarginRight(gap).Render(boxes[0]),
			boxes[1],
		)
		row1 := lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().MarginRight(gap).Render(boxes[2]),
			boxes[3],
		)
		cards = lipgloss.JoinVertical(lipgloss.Center, row0, lipgloss.NewStyle().Height(1).Render(""), row1)
	default:
		parts := make([]string, len(boxes))
		for i, b := range boxes {
			if i < len(boxes)-1 {
				parts[i] = lipgloss.NewStyle().MarginRight(gap).Render(b)
			} else {
				parts[i] = b
			}
		}
		cards = lipgloss.JoinHorizontal(lipgloss.Top, parts...)
	}

	hl := lipgloss.NewStyle().Foreground(m.theme.Highlight).Bold(true)
	mt := lipgloss.NewStyle().Foreground(m.theme.Muted)
	hint := lipgloss.JoinHorizontal(lipgloss.Center,
		hl.Render("←→↑↓"), mt.Render(" move  ·  "),
		hl.Render("1–4"), mt.Render(" pick  ·  "),
		hl.Render("enter"), mt.Render(" select  ·  "),
		hl.Render("q"), mt.Render(" quit  ·  "),
		hl.Render("t"), mt.Render(" compact"),
	)

	var header string
	if m.compact {
		header = renderCompactHeader("Choose how you'd like to measure your connection")
	} else {
		header = renderHeader("Choose how you'd like to measure your connection")
	}

	rule := lipgloss.NewStyle().Foreground(m.theme.Border).Render(strings.Repeat("─", 36))

	stack := lipgloss.JoinVertical(lipgloss.Center,
		header,
		rule,
		"",
		cards,
		"",
		hint,
	)

	ch := m.height
	if ch <= 0 {
		ch = 30
	}
	return apptheme.PaintScreen(m.theme, m.width, ch, stack)
}

func (m *menuModel) renderBox(i int, it menuItem, cardWidth int) string {
	selected := i == m.cursor || (m.hovered >= 0 && i == m.hovered)

	accent := m.theme.AccentDL
	fill := m.theme.MenuSelectDL
	switch it.screen {
	case screenMonitor:
		accent = m.theme.AccentUL
		fill = m.theme.MenuSelectUL
	case screenSettings:
		accent = m.theme.AccentLat
		fill = m.theme.MenuSelectSet
	case screenExit:
		accent = m.theme.AccentHL
		fill = m.theme.MenuSelectExit
	}

	var bg lipgloss.TerminalColor
	if selected {
		bg = fill
	} else {
		bg = m.theme.MenuIdleFill
	}

	innerW := cardWidth - 4
	if innerW < 12 {
		innerW = 12
	}

	cell := func(fg lipgloss.TerminalColor, bold bool) lipgloss.Style {
		s := lipgloss.NewStyle().Foreground(fg).Background(bg)
		if bold {
			s = s.Bold(true)
		}
		return s
	}
	space := lipgloss.NewStyle().Background(bg)
	line := func(parts ...string) string {
		joined := strings.Join(parts, "")
		return lipgloss.NewStyle().Width(innerW).Background(bg).Inline(true).Render(joined)
	}

	ink := lipgloss.Color("#0a0e14")
	var chip, titleBlock string
	if selected {
		if it.hotkey != "" {
			chip = lipgloss.NewStyle().Foreground(ink).Background(accent).Bold(true).Padding(0, 1).Render(it.hotkey)
		}
		titleBlock = lipgloss.NewStyle().Foreground(ink).Background(accent).Bold(true).Padding(0, 1).Render(it.title)
	} else {
		if it.hotkey != "" {
			chip = lipgloss.NewStyle().Foreground(accent).Background(bg).Bold(true).Padding(0, 1).Render(it.hotkey)
		}
		titleBlock = lipgloss.NewStyle().Foreground(accent).Background(bg).Bold(true).Padding(0, 1).Render(it.title)
	}
	titleRow := line(chip, space.Render(" "), titleBlock)

	subFG := m.theme.Muted
	if selected {
		subFG = m.theme.Foreground
	}
	subRow := line(space.Render("  "), cell(subFG, false).Render(it.subtitle))

	divCh := "─"
	if selected {
		divCh = "━"
	}
	div := line(cell(accent, false).Render(strings.Repeat(divCh, min(innerW, 20))))

	featRows := make([]string, 3)
	for j := 0; j < 3; j++ {
		if j < len(it.features) {
			bullet := cell(accent, false).Render("› ")
			if !selected {
				bullet = cell(m.theme.Border, false).Render("· ")
			}
			featRows[j] = line(space.Render(" "), bullet, cell(m.theme.Muted, false).Render(it.features[j]))
		} else {
			featRows[j] = line("")
		}
	}

	var badgeRow string
	if it.badge != "" {
		var badge string
		if selected {
			badge = lipgloss.NewStyle().Foreground(ink).Background(accent).Bold(true).Padding(0, 1).Render(it.badge)
		} else {
			badge = lipgloss.NewStyle().Foreground(accent).Background(bg).Bold(true).Render(" " + it.badge + " ")
		}
		badgeRow = line(space.Render(" "), badge)
	} else if selected {
		badgeRow = line(space.Render(" "), cell(accent, true).Render("↵ enter"))
	} else {
		badgeRow = line("")
	}

	topBar := line("")
	if selected {
		topBar = line(cell(accent, false).Render(strings.Repeat("▀", innerW)))
	}

	body := strings.Join([]string{
		topBar,
		titleRow,
		subRow,
		line(""),
		div,
		line(""),
		featRows[0],
		featRows[1],
		featRows[2],
		line(""),
		badgeRow,
		line(""), // keep TUNE / LIVE / etc. inside the selected fill
	}, "\n")

	var borderCol lipgloss.TerminalColor = m.theme.Border
	if selected {
		borderCol = accent
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderCol).
		Background(bg).
		Padding(1, 2).
		Width(cardWidth).
		Align(lipgloss.Left).
		Render(body)

	if selected {
		p := pulseFactor(m.pulse)
		gw := int(float64(cardWidth) * (0.72 + 0.28*p))
		if gw < cardWidth/2 {
			gw = cardWidth / 2
		}
		if gw > cardWidth {
			gw = cardWidth
		}
		bar := lipgloss.NewStyle().Foreground(accent).Bold(true).Render(strings.Repeat("▀", gw))
		pad := (cardWidth - gw) / 2
		if pad < 0 {
			pad = 0
		}
		under := strings.Repeat(" ", pad) + bar
		box = lipgloss.JoinVertical(lipgloss.Left, box, under)
	} else {
		box = lipgloss.JoinVertical(lipgloss.Left, box, strings.Repeat(" ", cardWidth))
	}
	return box
}

func pulseFactor(p float64) float64 {
	frac := p - float64(int(p))
	if frac < 0.5 {
		return 0.6 + frac*0.8
	}
	return 1.0 - (frac-0.5)*0.8
}

func (m *menuModel) applyTheme(t apptheme.Theme) {
	m.theme = t
	m.spinner.Style = lipgloss.NewStyle().Foreground(t.Highlight)
}
