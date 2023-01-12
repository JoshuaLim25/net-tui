// net-tui is a terminal-based network monitoring dashboard.
//
// It provides a unified interface for common networking tasks:
// viewing connections, listening ports, and interface statistics.
package main

import (
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	psnet "github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

func main() {
	p := tea.NewProgram(newModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// ----------------------------------------------------------------------------
// Messages
// ----------------------------------------------------------------------------

// triggers periodic data refresh
type tickMsg time.Time

// carries refreshed network data
type dataMsg struct {
	connections []connection
	ports       []port
	interfaces  []iface
}

// ----------------------------------------------------------------------------
// Model
// ----------------------------------------------------------------------------

// identifies which view is active.
type tab int

const (
	tabConnections tab = iota
	tabPorts
	tabInterfaces
)

func (t tab) String() string {
	return [...]string{"Connections", "Ports", "Interfaces"}[t]
}

// model holds all application state.
type model struct {
	tab    tab
	cursor int
	offset int
	width  int
	height int

	connections []connection
	ports       []port
	interfaces  []iface
}

func newModel() model {
	return model{tab: tabConnections}
}

// Init starts the initial data fetch and tick timer.
func (m model) Init() tea.Cmd {
	return tea.Batch(fetchData, tick())
}

// Update handles all messages and returns the updated model.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tickMsg:
		return m, tea.Batch(fetchData, tick())
	case dataMsg:
		m.connections = msg.connections
		m.ports = msg.ports
		m.interfaces = msg.interfaces
		m.clampCursor()
		return m, nil
	}
	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
	max := m.listLen() - 1
	if max < 0 {
		max = 0
	}
	if m.cursor > max {
		m.cursor = max
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

// View renders the entire UI.
func (m model) View() string {
	if m.width == 0 {
		return "loading..."
	}

	var b strings.Builder
	b.WriteString(m.viewHeader())
	b.WriteString(m.viewTabs())
	b.WriteString("\n")
	b.WriteString(m.viewContent())
	b.WriteString(m.viewFooter())
	return b.String()
}

func (m model) viewHeader() string {
	title := styles.title.Render(" net-tui ")
	clock := styles.dim.Render(time.Now().Format("15:04:05"))
	gap := max(m.width - lipgloss.Width(title) - lipgloss.Width(clock), 0)
	return title + strings.Repeat(" ", gap) + clock + "\n"
}

func (m model) viewTabs() string {
	var tabs []string
	for i := range 3 {
		t := tab(i)
		style := styles.tabInactive
		if t == m.tab {
			style = styles.tabActive
		}
		tabs = append(tabs, style.Render(" "+t.String()+" "))
	}
	return strings.Join(tabs, " ") + "\n"
}

func (m model) viewContent() string {
	height := m.height - 5 // header, tabs, footer
	if height < 1 {
		height = 1
	}

	switch m.tab {
	case tabConnections:
		return m.viewConnections(height)
	case tabPorts:
		return m.viewPorts(height)
	case tabInterfaces:
		return m.viewInterfaces(height)
	}
	return ""
}

func (m model) viewFooter() string {
	help := "q quit • tab/1-3 switch • j/k navigate"
	return "\n" + styles.dim.Render(help)
}

// ----------------------------------------------------------------------------
// Connection View
// ----------------------------------------------------------------------------

func (m *model) viewConnections(height int) string {
	var b strings.Builder

	header := fmt.Sprintf("%-7s %-21s %-21s %-11s %s",
		"PROTO", "LOCAL", "REMOTE", "STATE", "PROCESS")
	b.WriteString(styles.header.Render(header) + "\n")

	m.adjustOffset(height - 1)
	visible := m.visibleRange(len(m.connections), height-1)

	for i := visible.start; i < visible.end; i++ {
		c := m.connections[i]
		line := fmt.Sprintf("%-7s %-21s %-21s %-11s %s",
			c.proto,
			truncate(c.local, 21),
			truncate(c.remote, 21),
			c.state,
			truncate(c.process, 15),
		)
		if i == m.cursor {
			b.WriteString(styles.selected.Render(line))
		} else {
			b.WriteString(line)
		}
		b.WriteString("\n")
	}

	b.WriteString(styles.dim.Render(fmt.Sprintf("\n%d connections", len(m.connections))))
	return b.String()
}

// ----------------------------------------------------------------------------
// Ports View
// ----------------------------------------------------------------------------

func (m *model) viewPorts(height int) string {
	var b strings.Builder

	header := fmt.Sprintf("%-7s %-7s %-16s %-8s %s",
		"PORT", "PROTO", "ADDRESS", "PID", "PROCESS")
	b.WriteString(styles.header.Render(header) + "\n")

	m.adjustOffset(height - 1)
	visible := m.visibleRange(len(m.ports), height-1)

	for i := visible.start; i < visible.end; i++ {
		p := m.ports[i]
		addr := p.addr
		if addr == "" || addr == "0.0.0.0" || addr == "::" {
			addr = "*"
		}
		line := fmt.Sprintf("%-7d %-7s %-16s %-8d %s",
			p.port,
			p.proto,
			addr,
			p.pid,
			truncate(p.process, 20),
		)
		if i == m.cursor {
			b.WriteString(styles.selected.Render(line))
		} else {
			b.WriteString(line)
		}
		b.WriteString("\n")
	}

	b.WriteString(styles.dim.Render(fmt.Sprintf("\n%d listening ports", len(m.ports))))
	return b.String()
}

// ----------------------------------------------------------------------------
// Interfaces View
// ----------------------------------------------------------------------------

func (m *model) viewInterfaces(height int) string {
	var b strings.Builder

	header := fmt.Sprintf("%-12s %-6s %-22s %-12s %s",
		"NAME", "STATE", "ADDRESS", "RX", "TX")
	b.WriteString(styles.header.Render(header) + "\n")

	m.adjustOffset(height - 1)
	visible := m.visibleRange(len(m.interfaces), height-1)

	for i := visible.start; i < visible.end; i++ {
		ifc := m.interfaces[i]
		state := styles.stateDown.Render("down")
		if ifc.up {
			state = styles.stateUp.Render("up")
		}
		addr := "-"
		if len(ifc.addrs) > 0 {
			addr = ifc.addrs[0]
		}
		line := fmt.Sprintf("%-12s %-6s %-22s %-12s %s",
			ifc.name,
			state,
			truncate(addr, 22),
			formatBytes(ifc.rx),
			formatBytes(ifc.tx),
		)
		if i == m.cursor {
			b.WriteString(styles.selected.Render(line))
		} else {
			b.WriteString(line)
		}
		b.WriteString("\n")
	}

	b.WriteString(styles.dim.Render(fmt.Sprintf("\n%d interfaces", len(m.interfaces))))
	return b.String()
}

// ----------------------------------------------------------------------------
// Scroll helpers
// ----------------------------------------------------------------------------

type visibleRange struct{ start, end int }

func (m *model) adjustOffset(pageSize int) {
	if m.cursor < m.offset {
		m.offset = m.cursor
	} else if m.cursor >= m.offset+pageSize {
		m.offset = m.cursor - pageSize + 1
	}
}

func (m model) visibleRange(total, pageSize int) visibleRange {
	end := min(m.offset + pageSize, total)
	return visibleRange{m.offset, end}
}

// ----------------------------------------------------------------------------
// Data types
// ----------------------------------------------------------------------------

type connection struct {
	proto   string
	local   string
	remote  string
	state   string
	pid     int32
	process string
}

type port struct {
	port    uint32
	proto   string
	addr    string
	pid     int32
	process string
}

type iface struct {
	name  string
	up    bool
	addrs []string
	rx    uint64
	tx    uint64
}

// ----------------------------------------------------------------------------
// Data fetching
// ----------------------------------------------------------------------------

func tick() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func fetchData() tea.Msg {
	return dataMsg{
		connections: fetchConnections(),
		ports:       fetchPorts(),
		interfaces:  fetchInterfaces(),
	}
}

func fetchConnections() []connection {
	conns, err := psnet.Connections("all")
	if err != nil {
		return nil
	}

	var result []connection
	procCache := make(map[int32]string)

	for _, c := range conns {
		if c.Status == "" {
			continue
		}

		conn := connection{
			proto:  protoString(c.Type, c.Family),
			local:  formatAddr(c.Laddr.IP, c.Laddr.Port),
			remote: formatAddr(c.Raddr.IP, c.Raddr.Port),
			state:  c.Status,
			pid:    c.Pid,
		}

		if c.Pid > 0 {
			conn.process = getProcessName(c.Pid, procCache)
		}

		result = append(result, conn)
	}

	return result
}

func fetchPorts() []port {
	conns, err := psnet.Connections("all")
	if err != nil {
		return nil
	}

	seen := make(map[string]bool)
	var result []port
	procCache := make(map[int32]string)

	for _, c := range conns {
		if c.Status != "LISTEN" {
			continue
		}

		key := fmt.Sprintf("%d-%d", c.Laddr.Port, c.Type)
		if seen[key] {
			continue
		}
		seen[key] = true

		p := port{
			port:  c.Laddr.Port,
			proto: protoString(c.Type, c.Family),
			addr:  c.Laddr.IP,
			pid:   c.Pid,
		}

		if c.Pid > 0 {
			p.process = getProcessName(c.Pid, procCache)
		}

		result = append(result, p)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].port < result[j].port
	})

	return result
}

func fetchInterfaces() []iface {
	netIfaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	counters, _ := psnet.IOCounters(true)
	ioMap := make(map[string]psnet.IOCountersStat)
	for _, c := range counters {
		ioMap[c.Name] = c
	}

	var result []iface
	for _, ni := range netIfaces {
		// Skip loopback.
		if ni.Flags&net.FlagLoopback != 0 {
			continue
		}

		ifc := iface{
			name: ni.Name,
			up:   ni.Flags&net.FlagUp != 0,
		}

		if addrs, err := ni.Addrs(); err == nil {
			for _, a := range addrs {
				ifc.addrs = append(ifc.addrs, a.String())
			}
		}

		if io, ok := ioMap[ni.Name]; ok {
			ifc.rx = io.BytesRecv
			ifc.tx = io.BytesSent
		}

		result = append(result, ifc)
	}

	return result
}

func getProcessName(pid int32, cache map[int32]string) string {
	if name, ok := cache[pid]; ok {
		return name
	}
	name := ""
	if p, err := process.NewProcess(pid); err == nil {
		if n, err := p.Name(); err == nil {
			name = n
		}
	}
	cache[pid] = name
	return name
}

// ----------------------------------------------------------------------------
// Formatting helpers
// ----------------------------------------------------------------------------

func protoString(connType, family uint32) string {
	proto := "tcp"
	if connType == 2 {
		proto = "udp"
	}
	if family == 10 || family == 23 {
		proto += "6"
	}
	return proto
}

func formatAddr(ip string, port uint32) string {
	if ip == "" {
		ip = "*"
	}
	return fmt.Sprintf("%s:%d", ip, port)
}

func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

// ----------------------------------------------------------------------------
// Styles
// ----------------------------------------------------------------------------

var styles = struct {
	title       lipgloss.Style
	dim         lipgloss.Style
	header      lipgloss.Style
	selected    lipgloss.Style
	tabActive   lipgloss.Style
	tabInactive lipgloss.Style
	stateUp     lipgloss.Style
	stateDown   lipgloss.Style
}{
	title:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#88C0D0")),
	dim:         lipgloss.NewStyle().Foreground(lipgloss.Color("#4C566A")),
	header:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#81A1C1")),
	selected:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#2E3440")).Background(lipgloss.Color("#88C0D0")),
	tabActive:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#88C0D0")).Background(lipgloss.Color("#3B4252")),
	tabInactive: lipgloss.NewStyle().Foreground(lipgloss.Color("#4C566A")),
	stateUp:     lipgloss.NewStyle().Foreground(lipgloss.Color("#A3BE8C")),
	stateDown:   lipgloss.NewStyle().Foreground(lipgloss.Color("#BF616A")),
}
