package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"sort"

	tea "github.com/charmbracelet/bubbletea"
	psnet "github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

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

func fetchData() tea.Msg {
	conns, cErr := fetchConnections()
	ports, pErr := fetchPorts()
	ifaces, iErr := fetchInterfaces()

	return dataMsg{
		connections: conns,
		ports:       ports,
		interfaces:  ifaces,
		err:         errors.Join(cErr, pErr, iErr),
		privileged:  os.Geteuid() == 0,
	}
}

func fetchConnections() ([]connection, error) {
	conns, err := psnet.Connections("all")
	if err != nil {
		return nil, fmt.Errorf("fetch connections: %w", err)
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

	return result, nil
}

func fetchPorts() ([]port, error) {
	conns, err := psnet.Connections("all")
	if err != nil {
		return nil, fmt.Errorf("fetch ports: %w", err)
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

	return result, nil
}

func fetchInterfaces() ([]iface, error) {
	netIfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("fetch interfaces: %w", err)
	}

	counters, _ := psnet.IOCounters(true)
	ioMap := make(map[string]psnet.IOCountersStat)
	for _, c := range counters {
		ioMap[c.Name] = c
	}

	var result []iface
	for _, ni := range netIfaces {
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

	return result, nil
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
