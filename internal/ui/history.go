package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Foxemsx/riptide/internal/db"
	apptheme "github.com/Foxemsx/riptide/internal/theme"
)

// historyLimit is how many recent runs we always show.
const historyLimit = 10

// historySideMinWidth: terminal must be at least this wide to place history
// beside the speed-test card instead of below it.
const historySideMinWidth = 110

// historyBlock renders a polished "Recent tests" card for the speed-test view.
// Speeds are converted with the same unitMode as the main card (toggle with c).
func historyBlock(theme apptheme.Theme, runs []db.TestRun, width int, unit unitMode, hint string) string {
	if width < 28 {
		width = 28
	}
	bg := theme.MenuIdleFill
	ink := lipgloss.Color("#0a0e14")

	plain := lipgloss.NewStyle().Background(bg)
	fg := func(c lipgloss.TerminalColor, bold bool) lipgloss.Style {
		s := lipgloss.NewStyle().Foreground(c).Background(bg)
		if bold {
			s = s.Bold(true)
		}
		return s
	}
	innerW := width - 4
	if innerW < 24 {
		innerW = 24
	}
	line := func(parts ...string) string {
		return lipgloss.NewStyle().
			Width(innerW).
			Background(bg).
			Inline(true).
			Render(strings.Join(parts, ""))
	}

	titleChip := lipgloss.NewStyle().
		Foreground(ink).
		Background(theme.AccentLat).
		Bold(true).
		Padding(0, 1).
		Render("Recent tests")

	unitChip := lipgloss.NewStyle().
		Foreground(ink).
		Background(theme.AccentDL).
		Bold(true).
		Padding(0, 1).
		Render(unit.label())

	var body []string
	body = append(body, line(
		titleChip,
		plain.Render(" "),
		unitChip,
		plain.Render(" "),
		fg(theme.Muted, false).Render("latest 10"),
	))
	body = append(body, line(fg(theme.Border, false).Render(strings.Repeat("─", min(innerW, 48)))))

	// Column widths adapt to card width.
	nameW := 14
	rateW := 9
	if innerW >= 52 {
		nameW = 16
		rateW = 10
	}
	if innerW < 40 {
		nameW = 10
		rateW = 8
	}
	whenW := 11
	pingW := 6

	if len(runs) == 0 {
		body = append(body, line(fg(theme.Muted, false).Render("No runs yet — finish a test")))
		body = append(body, line(fg(theme.Muted, false).Render("or press s to name one.")))
	} else {
		// Header with unit-aware rate labels
		ulab := unit.short()
		body = append(body, line(
			fg(theme.Muted, true).Render(padRight("when", whenW)),
			fg(theme.Muted, true).Render(padRight("name", nameW)),
			fg(theme.Muted, true).Render(padLeft("↓"+ulab, rateW)),
			fg(theme.Muted, true).Render(padLeft("↑"+ulab, rateW)),
			fg(theme.Muted, true).Render(padLeft("ping", pingW)),
		))
		for i, r := range runs {
			if i >= historyLimit {
				break
			}
			when := padRight(db.FormatWhen(r.CreatedAt), whenW)
			name := padRight(truncate(r.Name, nameW-1), nameW)
			dl := padLeft(fmtSpeedUnit(r.DownloadMbps, unit), rateW)
			ul := padLeft(fmtSpeedUnit(r.UploadMbps, unit), rateW)
			pg := padLeft(fmt.Sprintf("%.0f", r.PingMs), pingW)
			nameStyle := fg(theme.Foreground, i == 0)
			body = append(body, line(
				fg(theme.Muted, false).Render(when),
				nameStyle.Render(name),
				fg(theme.Download, i == 0).Render(dl),
				fg(theme.Upload, i == 0).Render(ul),
				fg(theme.Latency, false).Render(pg),
			))
		}
	}

	if hint != "" {
		body = append(body, line(""))
		body = append(body, line(fg(theme.Muted, false).Render(hint)))
	} else {
		body = append(body, line(""))
		body = append(body, line(fg(theme.Muted, false).Render("c units  ·  s save")))
	}

	content := strings.Join(body, "\n")
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Border).
		Background(bg).
		Padding(0, 1).
		Width(width).
		Render(content)
}

// short is a compact unit label for tight column headers.
func (u unitMode) short() string {
	switch u {
	case unitKB:
		return "KB"
	case unitMB:
		return "MB"
	case unitGB:
		return "GB"
	default:
		return "Mb"
	}
}

// fmtSpeedUnit formats a Mbps value under the active unitMode for history rows.
func fmtSpeedUnit(mbps float64, u unitMode) string {
	if mbps <= 0 {
		return "—"
	}
	switch u {
	case unitKB:
		// 1 Mbps = 125 KB/s
		v := mbps * 125
		if v >= 10000 {
			return fmt.Sprintf("%.0fk", v/1000)
		}
		if v >= 100 {
			return fmt.Sprintf("%.0f", v)
		}
		return fmt.Sprintf("%.0f", v)
	case unitMB:
		// 1 Mbps = 0.125 MB/s
		v := mbps * 0.125
		if v >= 100 {
			return fmt.Sprintf("%.0f", v)
		}
		if v >= 10 {
			return fmt.Sprintf("%.1f", v)
		}
		return fmt.Sprintf("%.2f", v)
	case unitGB:
		v := mbps * 0.000125
		if v >= 1 {
			return fmt.Sprintf("%.2f", v)
		}
		return fmt.Sprintf("%.3f", v)
	default:
		// Mbps / Gbps auto
		if mbps >= 1000 {
			return fmt.Sprintf("%.1fG", mbps/1000)
		}
		if mbps >= 100 {
			return fmt.Sprintf("%.0f", mbps)
		}
		return fmt.Sprintf("%.1f", mbps)
	}
}

func fmtMbpsShort(mbps float64) string {
	return fmtSpeedUnit(mbps, unitAuto)
}

func padRight(s string, w int) string {
	vis := lipgloss.Width(s)
	if vis >= w {
		return truncate(s, w)
	}
	return s + strings.Repeat(" ", w-vis)
}

func padLeft(s string, w int) string {
	vis := lipgloss.Width(s)
	if vis >= w {
		return truncate(s, w)
	}
	return strings.Repeat(" ", w-vis) + s
}

// --- save name modal -----------------------------------------------------

// savePromptModel is a small centered modal for naming a run before save.
type savePromptModel struct {
	theme                        apptheme.Theme
	width                        int
	height                       int
	input                        textinput.Model
	kind                         string // speed
	dl, ul, ping, dlPeak, ulPeak float64
	server                       string
	active                       bool
}

func newSavePrompt(theme apptheme.Theme, kind string) savePromptModel {
	ti := textinput.New()
	ti.Placeholder = "Name this run…"
	ti.CharLimit = 48
	ti.Width = 40
	ti.Prompt = "› "
	return savePromptModel{theme: theme, input: ti, kind: kind}
}

func (s *savePromptModel) styleInput() {
	bg := s.theme.MenuIdleFill
	s.input.PromptStyle = lipgloss.NewStyle().Foreground(s.theme.AccentHL).Background(bg).Bold(true)
	s.input.TextStyle = lipgloss.NewStyle().Foreground(s.theme.Foreground).Background(bg)
	s.input.PlaceholderStyle = lipgloss.NewStyle().Foreground(s.theme.Muted).Background(bg)
	s.input.Cursor.Style = lipgloss.NewStyle().Foreground(s.theme.Download).Background(bg)
	s.input.Cursor.TextStyle = lipgloss.NewStyle().Foreground(s.theme.Foreground).Background(bg)
}

func (s *savePromptModel) open(dl, ul, ping, dlPeak, ulPeak float64, server string) {
	s.dl, s.ul, s.ping = dl, ul, ping
	s.dlPeak, s.ulPeak = dlPeak, ulPeak
	s.server = server
	s.styleInput()
	s.input.SetValue(db.AutoName(s.kind, time.Now()))
	s.input.CursorEnd()
	s.input.Focus()
	s.active = true
}

func (s *savePromptModel) close() {
	s.active = false
	s.input.Blur()
}

// saveRunMsg is emitted when the user confirms a named save.
type saveRunMsg struct {
	run db.TestRun
}

func (s *savePromptModel) Update(msg tea.Msg) tea.Cmd {
	if !s.active {
		return nil
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			s.close()
			return nil
		case "enter":
			name := strings.TrimSpace(s.input.Value())
			if name == "" {
				name = db.AutoName(s.kind, time.Now())
			}
			run := db.TestRun{
				Name:         name,
				Kind:         s.kind,
				DownloadMbps: s.dl,
				UploadMbps:   s.ul,
				PingMs:       s.ping,
				DownloadPeak: s.dlPeak,
				UploadPeak:   s.ulPeak,
				Server:       s.server,
				CreatedAt:    time.Now(),
			}
			s.close()
			return func() tea.Msg { return saveRunMsg{run: run} }
		}
	}
	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)
	return cmd
}

func (s *savePromptModel) View() string {
	if !s.active {
		return ""
	}
	s.styleInput()

	bg := s.theme.MenuIdleFill
	ink := lipgloss.Color("#0a0e14")
	const innerW = 44

	line := func(parts ...string) string {
		return lipgloss.NewStyle().
			Width(innerW).
			Background(bg).
			Inline(true).
			Render(strings.Join(parts, ""))
	}
	fg := func(c lipgloss.TerminalColor, bold bool) lipgloss.Style {
		st := lipgloss.NewStyle().Foreground(c).Background(bg)
		if bold {
			st = st.Bold(true)
		}
		return st
	}

	titleChip := lipgloss.NewStyle().
		Foreground(ink).Background(s.theme.AccentHL).Bold(true).Padding(0, 1).
		Render("Save test run")

	s.input.Width = innerW - 2
	if s.input.Width < 8 {
		s.input.Width = 8
	}
	inputView := lipgloss.NewStyle().Background(bg).Render(s.input.View())
	inputLine := lipgloss.PlaceHorizontal(
		innerW, lipgloss.Left, inputView,
		lipgloss.WithWhitespaceBackground(bg),
	)

	body := strings.Join([]string{
		line(titleChip),
		line(""),
		line(fg(s.theme.Muted, false).Render("Name it, then enter to store · esc to cancel")),
		line(""),
		line(
			fg(s.theme.Download, true).Render(fmt.Sprintf("↓ %s", fmtMbpsShort(s.dl))),
			fg(s.theme.Muted, false).Render("   "),
			fg(s.theme.Upload, true).Render(fmt.Sprintf("↑ %s", fmtMbpsShort(s.ul))),
			fg(s.theme.Muted, false).Render("   "),
			fg(s.theme.Latency, true).Render(fmt.Sprintf("◷ %.0f ms", s.ping)),
		),
		line(""),
		inputLine,
		line(""),
		line(fg(s.theme.Muted, false).Render("enter save  ·  esc cancel")),
	}, "\n")

	panel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(s.theme.Highlight).
		Background(bg).
		Padding(1, 2).
		Width(innerW + 4).
		Render(body)

	return apptheme.PaintScreen(s.theme, s.width, s.height, panel)
}
