# net-tui — Design Doc

## What it does
- **net-tui** is a terminal-based network monitoring dashboard. It aims to provide a unified interface for common networking tasks that typically require multiple tools (think `netstat`, `ss`, `lsof`, `nethogs`, `iftop`, etc.).
- In particular, it tries to eliminate a lot of the overhead around remembering random incarnations and one liners, allowing you to view the information you care about at a glance.

### Design goals
1. Single binary with opinionated config; to "just work" out of the box
2. Real-time, much like how tools like `htop` work. It tries to be relatively lean and just do the networking aspect really well
3. Simple, intuitive navigation and ui

### Who this is for
- Developers debugging local services or monitoring servers
- Specifically to alleviate mental overhead of remembering various bespoke commands
- Basically anyone who's tired of remembering `lsof -i :8080` vs `netstat -tulnp` or wants a cleaner view for these

---

## Installation

```bash
git clone https://github.com/josh/net-tui
cd net-tui
# Install system-wide to /usr/local/bin
make install
```


## Basic support

| Feature | Desc |
|---------|-------------|
| Connections view | Live TCP/UDP connections with process info |
| Ports view | Listening ports grouped by process |
| Interfaces view | Network interface stats (RX/TX bytes) |
| Tab navigation | Switch between views |
| Cursor navigation | Vim-style j/k movement with scrolling |
| Auto-refresh | 2-second polling interval |

