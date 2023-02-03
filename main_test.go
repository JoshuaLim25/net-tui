package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name  string
		input uint64
		want  string
	}{
		{"zero", 0, "0 B"},
		{"KB", 1024, "1.0 KB"},
		{"MB", 1024 * 1024, "1.0 MB"},
		{"GB", 1024 * 1024 * 1024, "1.0 GB"},
		{"UnderKB", 1023, "1023 B"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatBytes(tt.input); got != tt.want {
				t.Errorf("formatBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name  string
		input string
		max   int
		want  string
	}{
		{"no change", "hello", 10, "hello"},
		{"exact", "hello", 5, "hello"},
		{"truncate", "hello world", 8, "hello..."},
		{"short", "abc", 2, "ab"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := truncate(tt.input, tt.max); got != tt.want {
				t.Errorf("truncate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModelTabNavigation(t *testing.T) {
	m := newModel()

	// Initial state
	if m.tab != tabConnections {
		t.Errorf("initial tab should be connections, got %v", m.tab)
	}

	// Test tab switching
	msg := tea.KeyMsg{Type: tea.KeyTab}
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(model)

	if m.tab != tabPorts {
		t.Errorf("tab after 1 press should be ports, got %v", m.tab)
	}

	// Test cycling back to first tab
	msg = tea.KeyMsg{Type: tea.KeyTab}
	updatedModel, _ = m.Update(msg) // Ports -> Interfaces
	m = updatedModel.(model)
	updatedModel, _ = m.Update(msg) // Interfaces -> Connections
	m = updatedModel.(model)

	if m.tab != tabConnections {
		t.Errorf("tab after cycling should be connections, got %v", m.tab)
	}
}
