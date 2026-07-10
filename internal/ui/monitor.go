package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/Foxemsx/riptide/internal/engine"
	apptheme "github.com/Foxemsx/riptide/internal/theme"
)

// monitorModel is the live Bandwidth Monitor card.
type monitorModel struct {
	*cardState

	paused    bool
	startTime time.Time

	dlPeak float64
	ulPeak float64

	pingDone bool
}

func newMonitorModel(cs *cardState) *monitorModel {
	m := &monitorModel{cardState: cs}
	m.startTime = time.Now()
	return m
}

func (m *monitorModel) Start() tea.Cmd {
	bridgeLaunch(m.ctx, m.progress, m.events, func() {
		engine.RunMonitor(m.ctx, m.progress, tickInterval)
	})
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg { return tickMsg{} },
		listenCmd(m.events),
	)
}

func (m *monitorModel) reset() tea.Cmd {
	if m.cancel != nil {
		m.cancel()
	}
	w, h := m.width, m.height
	cs := newCardState(m.theme, m.compact)
	m.cardState = cs
	m.width, m.height = w, h
	m.syncLayout()
	m.startTime = time.Now()
	m.dlPeak = 0
	m.ulPeak = 0
	m.pingDone = false
	m.paused = false

	bridgeLaunch(m.ctx, m.progress, m.events, func() {
		engine.RunMonitor(m.ctx, m.progress, tickInterval)
	})
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg { return tickMsg{} },
		listenCmd(m.events),
	)
}

func (m *monitorModel) Update(msg tea.Msg) (tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			if m.cancel != nil {
				m.cancel()
			}
			return tea.Quit, false
		case "esc", "m":
			if m.cancel != nil {
				m.cancel()
			}
			return backToMenuCmd(), false
		case "?":
			m.showHelp = !m.showHelp
			return nil, false
		case "r":
			return m.reset(), false
		case "c":
			m.unit = (m.unit + 1) % 4
			return nil, false
		case "p":
			m.paused = !m.paused
			return nil, false
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.syncLayout()
		return nil, false

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return cmd, false

	case phaseMsg:
		m.phase = msg.phase
		if msg.phase == engine.PhaseConnected {
			if m.progress.ServerName != "" {
				m.serverName = m.progress.ServerName
			}
			m.phase = engine.PhaseUpload
		}
		return listenCmd(m.events), false

	case sampleMsg:
		if !m.paused {
			mbps := engine.BytesPerSecToMbps(msg.sample.Rate)
			switch msg.sample.Phase {
			case engine.PhaseDownload:
				m.dlTarget = mbps
				if mbps > m.dlPeak {
					m.dlPeak = mbps
				}
			case engine.PhaseUpload:
				m.ulTarget = mbps
				if mbps > m.ulPeak {
					m.ulPeak = mbps
				}
			}
		}
		return listenCmd(m.events), false

	case pingMsg:
		m.pingDisp = msg.ms
		m.pingDone = true
		return nil, false

	case tickMsg:
		m.advance()
		return m.tickCmd(), false
	}
	return nil, false
}

func (m *monitorModel) advance() {
	if m.paused {
		return
	}
	m.dlDisplay = lerp(m.dlDisplay, m.dlTarget, animFactor)
	m.ulDisplay = lerp(m.ulDisplay, m.ulTarget, animFactor)
	m.dlGraph.push(m.dlDisplay)
	m.ulGraph.push(m.ulDisplay)
}

func (m *monitorModel) View() string {
	m.syncLayout()

	var body strings.Builder

	if m.serverName != "" {
		inner := m.cardWidthFor() - 4
		body.WriteString(center(lipgloss.NewStyle().
			Foreground(m.theme.Muted).
			Render("watching "+m.serverName), inner))
		body.WriteString("\n\n")
	}

	modeLabel := lipgloss.NewStyle().Foreground(m.theme.Highlight).Bold(true).Render("● LIVE")
	if m.paused {
		modeLabel = lipgloss.NewStyle().Foreground(m.theme.Muted).Bold(true).Render("Ⅱ PAUSED")
	}
	body.WriteString(center(lipgloss.JoinHorizontal(lipgloss.Left, m.spinner.View()+" ", modeLabel), m.cardWidthFor()))
	body.WriteString("\n\n")

	body.WriteString(m.metricBlock(
		"↓ download", m.theme.Download, m.dlDisplay, m.dlGraph, m.dlPeak, engine.PhaseDownload,
	))
	body.WriteString("\n\n")

	body.WriteString(m.metricBlock(
		"↑ upload", m.theme.Upload, m.ulDisplay, m.ulGraph, m.ulPeak, engine.PhaseUpload,
	))
	body.WriteString("\n\n")

	uptime := time.Since(m.startTime).Round(time.Second)
	pingStr := "-"
	if m.pingDone {
		pingStr = fmt.Sprintf("%.0f ms", m.pingDisp)
	}
	left := lipgloss.NewStyle().Foreground(m.theme.Muted).Render("uptime " + uptime.String())
	right := lipgloss.NewStyle().Foreground(m.theme.Muted).Render(m.unit.label() + " · ping " + pingStr)
	body.WriteString(center(lipgloss.JoinHorizontal(lipgloss.Left, left, "    ", right), m.cardWidthFor()))

	hl := lipgloss.NewStyle().Foreground(m.theme.Highlight).Bold(true)
	mt := lipgloss.NewStyle().Foreground(m.theme.Muted)
	hint := lipgloss.JoinHorizontal(lipgloss.Center,
		hl.Render("esc"), mt.Render(" menu  ·  "),
		hl.Render("c"), mt.Render(" units  ·  "),
		hl.Render("p"), mt.Render(" pause  ·  "),
		hl.Render("r"), mt.Render(" reset  ·  "),
		hl.Render("?"), mt.Render(" help"),
	)
	body.WriteString("\n\n")
	body.WriteString(center(hint, m.cardWidthFor()))

	card := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Border).
		Background(m.theme.AppBg).
		Padding(1, 2).
		Width(m.cardWidthFor()).
		Render(body.String())

	var header string
	if m.compact {
		header = renderCompactHeader("Watching your connection in real time")
	} else {
		header = renderHeader("Watching your connection in real time")
	}
	stack := lipgloss.JoinVertical(lipgloss.Center,
		header,
		"",
		card,
	)

	if m.showHelp {
		return m.renderHelp()
	}

	return apptheme.PaintScreen(m.theme, m.width, m.height, stack)
}

func (m *monitorModel) renderHelp() string {
	return renderHelpPanel(m.theme, "Bandwidth — Help", []helpBinding{
		{keys: "esc / m", action: "back to main menu"},
		{keys: "?", action: "close this help"},
		{keys: "q", action: "quit riptide"},
		{keys: "p", action: "pause / resume monitoring"},
		{keys: "r", action: "restart the monitor"},
		{keys: "c", action: "cycle units  Mbps · KB/s · MB/s · GB/s"},
		{keys: "t", action: "toggle compact logo"},
	}, m.width, m.height)
}
