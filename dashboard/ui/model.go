package ui

import (
	"fmt"
	"sort"

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

	onlineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Bold(true)

	offlineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Bold(true)

	rowStyle = lipgloss.NewStyle().
			PaddingLeft(1)

	tableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Underline(true).
				PaddingBottom(1)

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 2)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))
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

	header := headerStyle.Render("Infrastructure Monitor")

	// Maps are unordered — without this the table jumps around every render.
	ids := make([]string, 0, len(m.snapshot))
	for id := range m.snapshot {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	tableHeader := tableHeaderStyle.Render(
		fmt.Sprintf("%-20s %-10s %8s %8s %8s %15s",
			"Machine", "Status", "CPU", "RAM", "Disk", "Last Seen"),
	)

	rows := ""
	for _, id := range ids {
		rows += renderRow(id, m.snapshot[id])
	}

	table := borderStyle.Render(tableHeader + "\n" + rows)

	help := dimStyle.Render("  q to quit")

	return fmt.Sprintf("%s\n%s\n%s\n", header, table, help)
}

func renderRow(id string, state metrics.AgentState) string {
	status := onlineStyle.Render("ONLINE ")
	cpu := fmt.Sprintf("%6.1f%%", state.Payload.CPU)
	ram := fmt.Sprintf("%6.1f%%", state.Payload.RAM)
	disk := fmt.Sprintf("%6.1f%%", state.Payload.Disk)
	lastSeen := ws.LastSeen(state.Payload.Timestamp)

	if !state.Online {
		status = offlineStyle.Render("OFFLINE")
		cpu = dimStyle.Render("   ---")
		ram = dimStyle.Render("   ---")
		disk = dimStyle.Render("   ---")
	}

	row := fmt.Sprintf("%-20s %-10s %8s %8s %8s %15s",
		id, status, cpu, ram, disk, lastSeen)

	return rowStyle.Render(row) + "\n"
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
