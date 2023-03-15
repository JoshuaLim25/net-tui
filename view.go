package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var styles = struct {
	title       lipgloss.Style
	dim         lipgloss.Style
	header      lipgloss.Style
	selected    lipgloss.Style
	tabActive   lipgloss.Style
	tabInactive lipgloss.Style
	stateUp     lipgloss.Style
	stateDown   lipgloss.Style
	error       lipgloss.Style
}{
	title:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6")),
	dim:         lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
	header:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4")),
	selected:    lipgloss.NewStyle().Reverse(true),
	tabActive:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6")).Underline(true),
	tabInactive: lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
	stateUp:     lipgloss.NewStyle().Foreground(lipgloss.Color("2")),
	stateDown:   lipgloss.NewStyle().Foreground(lipgloss.Color("1")),
	error:       lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true),
}

type column struct {
	title string
	width int
}

func (m model) View() string {
	if m.width == 0 {
		return " (loading...) "
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
	return styles.title.Render(" net-tui ") + "\n"
}

func (m model) viewTabs() string {
	numTabs := 3
	tabWidth := m.width / numTabs

	var tabs []string
	for i := 0; i < numTabs; i++ {
		t := tab(i)
		style := styles.tabInactive.Width(tabWidth).Align(lipgloss.Center)
		if t == m.tab {
			style = styles.tabActive.Width(tabWidth).Align(lipgloss.Center)
		}
		tabs = append(tabs, style.Render(t.String()))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...) + "\n"
}

func (m model) viewContent() string {
	switch m.tab {
	case tabConnections:
		cols := []column{
			{title: "PROTO", width: 7},
			{title: "LOCAL", width: 21},
			{title: "REMOTE", width: 21},
			{title: "STATE", width: 11},
			{title: "PROCESS", width: 0},
		}
		return m.renderTable(cols, len(m.connections), func(i int) string {
			c := m.connections[i]
			return fmt.Sprintf("%-7s %-21s %-21s %-11s %s",
				c.proto, truncate(c.local, 21), truncate(c.remote, 21), c.state, c.process)
		})
	case tabPorts:
		cols := []column{
			{title: "PORT", width: 7},
			{title: "PROTO", width: 7},
			{title: "ADDRESS", width: 16},
			{title: "PID", width: 8},
			{title: "PROCESS", width: 0},
		}
		return m.renderTable(cols, len(m.ports), func(i int) string {
			p := m.ports[i]
			addr := p.addr
			if addr == "" || addr == "0.0.0.0" || addr == "::" {
				addr = "*"
			}
			return fmt.Sprintf("%-7d %-7s %-16s %-8d %s",
				p.port, p.proto, truncate(addr, 16), p.pid, p.process)
		})
	case tabInterfaces:
		cols := []column{
			{title: "NAME", width: 12},
			{title: "STATE", width: 6},
			{title: "ADDRESS", width: 22},
			{title: "RX", width: 12},
			{title: "TX", width: 0},
		}
		return m.renderTable(cols, len(m.interfaces), func(i int) string {
			ifc := m.interfaces[i]
			state := "down"
			if ifc.up {
				state = "up"
			}
			addr := "-"
			if len(ifc.addrs) > 0 {
				addr = ifc.addrs[0]
			}
			return fmt.Sprintf("%-12s %-6s %-22s %-12s %s",
				ifc.name, state, truncate(addr, 22), formatBytes(ifc.rx), formatBytes(ifc.tx))
		})
	}
	return ""
}

func (m model) viewFooter() string {
	var b strings.Builder
	b.WriteString("\n")
	help := "q quit • tab/1-3 switch • j/k navigate"
	b.WriteString(styles.dim.Render(help))
	return b.String()
}

func (m model) renderTable(cols []column, rowCount int, renderRow func(int) string) string {
	if m.height < 6 {
		return " (terminal too small) "
	}

	var b strings.Builder

	// 1. Render Header
	var headerParts []string
	for i, col := range cols {
		w := col.width
		if i == len(cols)-1 && w == 0 {
			used := 0
			for j := 0; j < len(cols)-1; j++ {
				used += cols[j].width + 1
			}
			w = max(m.width-used, 0)
		}
		headerParts = append(headerParts, fmt.Sprintf("%-*s", w, col.title))
	}
	header := strings.Join(headerParts, " ")
	b.WriteString(styles.header.Render(header) + "\n")

	// 2. Handle Empty State
	if rowCount == 0 {
		msg := "  No data found."
		if m.err != nil {
			msg = "  Error: " + m.err.Error()
		}
		b.WriteString("\n" + styles.dim.Render(msg) + "\n")
		return b.String()
	}

	// 3. Render Rows
	pageSize := m.height - 6
	if pageSize < 1 {
		pageSize = 1
	}

	visible := m.visibleRange(rowCount, pageSize)

	for i := visible.start; i < visible.end; i++ {
		line := renderRow(i)
		if i == m.cursor {
			b.WriteString(styles.selected.Render(truncate(line, m.width)))
		} else {
			b.WriteString(truncate(line, m.width))
		}
		b.WriteString("\n")
	}

	return b.String()
}

type visibleRange struct{ start, end int }

func (m model) visibleRange(total, pageSize int) visibleRange {
	end := min(m.offset+pageSize, total)
	return visibleRange{m.offset, end}
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

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
