package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

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
	height := m.height - 5
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
	var b strings.Builder
	b.WriteString("\n")
	if m.err != nil {
		b.WriteString(styles.error.Render("error: " + m.err.Error()))
		b.WriteString("\n")
	}
	help := "q quit • tab/1-3 switch • j/k navigate"
	b.WriteString(styles.dim.Render(help))
	return b.String()
}

func (m *model) viewConnections(height int) string {
	var b strings.Builder

	protoWidth := 7
	stateWidth := 11

	spaceAvailable := m.width - protoWidth - stateWidth - 4 // spacing
	if spaceAvailable < 0 {
		spaceAvailable = 0
	}

	addrW := 21
	if spaceAvailable < addrW*2+10 {
		addrW = spaceAvailable / 3
	}
	processW := max(spaceAvailable - (addrW * 2), 10)

	header := fmt.Sprintf("%-*s %-*s %-*s %-*s %s",
		protoWidth, "PROTO", addrW, "LOCAL", addrW, "REMOTE", stateWidth, "STATE", "PROCESS")
	b.WriteString(styles.header.Render(header) + "\n")

	if len(m.connections) == 0 {
		b.WriteString("\n" + styles.dim.Render("  No active connections found.") + "\n")
		return b.String()
	}

	m.adjustOffset(height - 1)
	visible := m.visibleRange(len(m.connections), height-1)

	for i := visible.start; i < visible.end; i++ {
		c := m.connections[i]

		stateStr := c.state
		stateStyle := lipgloss.NewStyle()
		switch c.state {
		case "ESTABLISHED":
			stateStyle = styles.stateGreen
		case "TIME_WAIT", "CLOSE_WAIT":
			stateStyle = styles.stateYellow
		case "LISTEN":
			stateStyle = styles.stateUp
		}

		line := fmt.Sprintf("%-*s %-*s %-*s %-*s %s",
			protoWidth, c.proto,
			addrW, truncate(c.local, addrW),
			addrW, truncate(c.remote, addrW),
			stateWidth, stateStyle.Render(truncate(stateStr, stateWidth)),
			truncate(c.process, processW),
		)

		if i == m.cursor {
			line = fmt.Sprintf("%-*s %-*s %-*s %-*s %s",
				protoWidth, c.proto,
				addrW, truncate(c.local, addrW),
				addrW, truncate(c.remote, addrW),
				stateWidth, truncate(stateStr, stateWidth),
				truncate(c.process, processW),
			)
			b.WriteString(styles.selected.Render(line))
		} else {
			b.WriteString(line)
		}
		b.WriteString("\n")
	}

	b.WriteString(styles.dim.Render(fmt.Sprintf("\n%d connections", len(m.connections))))
	return b.String()
}

func (m *model) viewPorts(height int) string {
	var b strings.Builder

	header := fmt.Sprintf("%-7s %-7s %-16s %-8s %s",
		"PORT", "PROTO", "ADDRESS", "PID", "PROCESS")
	b.WriteString(styles.header.Render(header) + "\n")

	if len(m.ports) == 0 {
		b.WriteString("\n" + styles.dim.Render("  No listening ports found.") + "\n")
		return b.String()
	}

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

func (m *model) viewInterfaces(height int) string {
	var b strings.Builder

	header := fmt.Sprintf("%-12s %-6s %-22s %-12s %s",
		"NAME", "STATE", "ADDRESS", "RX", "TX")
	b.WriteString(styles.header.Render(header) + "\n")

	if len(m.interfaces) == 0 {
		b.WriteString("\n" + styles.dim.Render("  No network interfaces found.") + "\n")
		return b.String()
	}

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
