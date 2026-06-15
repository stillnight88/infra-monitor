package ui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stillnight88/infra-monitor/dashboard/ws"
	"github.com/stillnight88/infra-monitor/server/metrics"
)

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")).
			PaddingBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			PaddingBottom(1)

	onlineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Bold(true)

	offlineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Bold(true)

	tableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Underline(true).
				PaddingBottom(1)

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("5")).
			Padding(1, 2)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	barNormalStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	barWarningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	barDangerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)

// SnapshotReceived is a Bubble Tea message carrying a new snapshot.
type SnapshotReceived struct {
	Snapshot ws.SnapshotMsg
}

type ConnectionLost struct{}

// Model is the Bubble Tea model, holds all state the UI needs to render.
type Model struct {
	snapshot ws.SnapshotMsg
	ch       <-chan ws.SnapshotMsg
	quitting bool
	err      string
}

// New returns an initialised Model connected to the snapshot channel.
func New(ch <-chan ws.SnapshotMsg) Model {
	return Model{
		snapshot: make(ws.SnapshotMsg),
		ch:       ch,
	}
}

// Init returns the first Cmd to run when the program starts.
func (m Model) Init() tea.Cmd {
	return waitForSnapshot(m.ch)
}

// Update handles incoming messages and returns updated model + next Cmd.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case SnapshotReceived:
		m.snapshot = msg.Snapshot
		return m, waitForSnapshot(m.ch)

	case ConnectionLost:
		m.err = "connection to server lost"
		return m, tea.Quit

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

// View renders the current model state as a string.
func (m Model) View() string {
	if m.quitting {
		return "goodbye.\n"
	}

	if m.err != "" {
		return fmt.Sprintf("error: %s\n", m.err)
	}

	if len(m.snapshot) == 0 {
		return dimStyle.Render("waiting for agents...") + "\n"
	}

	// Maps are unordered — without this the table jumps around every render.
	ids := make([]string, 0, len(m.snapshot))
	for id := range m.snapshot {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	online := 0
	for _, state := range m.snapshot {
		if state.Online {
			online++
		}
	}

	header := headerStyle.Render("Infrastructure Monitor")
	subtitle := subtitleStyle.Render(
		fmt.Sprintf("%d online / %d total", online, len(m.snapshot)),
	)

	tableHeader := tableHeaderStyle.Render(
		fmt.Sprintf("%-22s %-10s %-16s %-16s %-16s %s",
			"Machine", "Status", "CPU", "RAM", "Disk", "Last Seen"),
	)

	rows := ""
	for _, id := range ids {
		rows += renderRow(id, m.snapshot[id])
	}

	table := borderStyle.Render(tableHeader + "\n" + rows)

	help := dimStyle.Render("  q to quit")

	return fmt.Sprintf("%s\n%s\n%s\n%s\n", header, subtitle, table, help)
}

func renderRow(id string, state metrics.AgentState) string {
	lastSeen := ws.LastSeen(state.Payload.Timestamp)

	var status, cpu, ram, disk, seen string

	if state.Online {
		status = onlineStyle.Render(fmt.Sprintf("%-8s", "ONLINE"))
		cpu = fmt.Sprintf("%5.1f%% ", state.Payload.CPU) + renderBar(state.Payload.CPU)
		ram = fmt.Sprintf("%5.1f%% ", state.Payload.RAM) + renderBar(state.Payload.RAM)
		disk = fmt.Sprintf("%5.1f%% ", state.Payload.Disk) + renderBar(state.Payload.Disk)
		seen = fmt.Sprintf("%10s", lastSeen)
	} else {
		status = offlineStyle.Render(fmt.Sprintf("%-8s", "OFFLINE"))
		cpu = dimStyle.Render(fmt.Sprintf("%14s", "---"))
		ram = dimStyle.Render(fmt.Sprintf("%14s", "---"))
		disk = dimStyle.Render(fmt.Sprintf("%14s", "---"))
		seen = dimStyle.Render(fmt.Sprintf("%10s", lastSeen))
	}

	return fmt.Sprintf("%-22s %s   %s   %s   %s   %s\n",
		truncate(id, 20), status, cpu, ram, disk, seen)
}

// renderBar returns a 8-char visual bar for a percentage value, Green below 60, yellow 60-80, red above 80.
func renderBar(pct float64) string {
	const width = 6
	filled := int((pct / 100.0) * width)
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)

	switch {
	case pct >= 80:
		return barDangerStyle.Render(bar)
	case pct >= 60:
		return barWarningStyle.Render(bar)
	default:
		return barNormalStyle.Render(bar)
	}
}

// truncate shortens a string to max length, adding … if cut.
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func waitForSnapshot(ch <-chan ws.SnapshotMsg) tea.Cmd {
	return func() tea.Msg {
		snapshot, ok := <-ch
		if !ok {
			return ConnectionLost{}
		}
		return SnapshotReceived{Snapshot: snapshot}
	}
}
