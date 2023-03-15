package main

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type tickMsg time.Time

type dataMsg struct {
	connections []connection
	ports       []port
	interfaces  []iface
	err         error
	privileged  bool
}

type tab int

const (
	tabConnections tab = iota
	tabPorts
	tabInterfaces
)

func (t tab) String() string {
	return [...]string{"Connections", "Ports", "Interfaces"}[t]
}

type model struct {
	tab    tab
	cursor int
	offset int
	width  int
	height int
	err    error

	connections []connection
	ports       []port
	interfaces  []iface
}

func newModel() model {
	return model{
		tab: tabConnections,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(fetchData, tick())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m, cmd := m.handleKey(msg)
		m.adjustOffset()
		return m, cmd
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.adjustOffset()
		return m, nil
	case tickMsg:
		return m, tea.Batch(fetchData, tick())
	case dataMsg:
		m.connections = msg.connections
		m.ports = msg.ports
		m.interfaces = msg.interfaces
		m.err = msg.err
		m.clampCursor()
		m.adjustOffset()
		return m, nil
	}
	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "tab", "l", "right":
		m.tab = (m.tab + 1) % 3
		m.cursor, m.offset = 0, 0
	case "shift+tab", "h", "left":
		m.tab = (m.tab + 2) % 3
		m.cursor, m.offset = 0, 0
	case "j", "down":
		m.cursor++
		m.clampCursor()
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "g", "home":
		m.cursor, m.offset = 0, 0
	case "G", "end":
		m.cursor = m.listLen() - 1
		m.clampCursor()
	case "1":
		m.tab, m.cursor, m.offset = tabConnections, 0, 0
	case "2":
		m.tab, m.cursor, m.offset = tabPorts, 0, 0
	case "3":
		m.tab, m.cursor, m.offset = tabInterfaces, 0, 0
	}
	return m, nil
}

func (m *model) clampCursor() {
	maxIdx := m.listLen() - 1
	if maxIdx < 0 {
		maxIdx = 0
	}
	if m.cursor > maxIdx {
		m.cursor = maxIdx
	}
}

func (m model) listLen() int {
	switch m.tab {
	case tabConnections:
		return len(m.connections)
	case tabPorts:
		return len(m.ports)
	case tabInterfaces:
		return len(m.interfaces)
	}
	return 0
}

func (m *model) adjustOffset() {
	pageSize := m.height - 6 // header, tabs, footer, table header
	if pageSize < 1 {
		pageSize = 1
	}

	if m.cursor < m.offset {
		m.offset = m.cursor
	} else if m.cursor >= m.offset+pageSize {
		m.offset = m.cursor - pageSize + 1
	}
}

func tick() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
