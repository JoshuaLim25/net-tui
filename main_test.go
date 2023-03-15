package main

import (
	"testing"

	"github.com/bradleyjkemp/cupaloy/v2"
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

func TestViewConnectionsSnapshot(t *testing.T) {
	m := newModel()
	m.width = 80
	m.height = 24
	m.connections = []connection{
		{proto: "tcp", local: "127.0.0.1:8080", remote: "0.0.0.0:0", state: "LISTEN", process: "nginx"},
		{proto: "tcp", local: "192.168.1.10:443", remote: "1.2.3.4:5678", state: "ESTABLISHED", process: "chrome"},
	}

	cupaloy.SnapshotT(t, m.View())
}

func TestViewPortsSnapshot(t *testing.T) {
	m := newModel()
	m.width = 80
	m.height = 24
	m.tab = tabPorts
	m.ports = []port{
		{port: 80, proto: "tcp", addr: "0.0.0.0", pid: 1234, process: "httpd"},
	}

	cupaloy.SnapshotT(t, m.View())
}
