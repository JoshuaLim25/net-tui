package main

import "github.com/charmbracelet/lipgloss"

var styles = struct {
	title       lipgloss.Style
	dim         lipgloss.Style
	header      lipgloss.Style
	selected    lipgloss.Style
	tabActive   lipgloss.Style
	tabInactive lipgloss.Style
	stateUp     lipgloss.Style
	stateDown   lipgloss.Style
	stateGreen  lipgloss.Style
	stateYellow lipgloss.Style
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
	stateGreen:  lipgloss.NewStyle().Foreground(lipgloss.Color("2")),
	stateYellow: lipgloss.NewStyle().Foreground(lipgloss.Color("3")),
	error:       lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true),
}
