# net-tui — Design Doc

## What it does
- **netscan** is a terminal-based network monitoring dashboard. It aims to provide a unified interface for common networking tasks that typically require multiple tools (think `netstat`, `ss`, `lsof`, `nethogs`, `iftop`, etc.).
- In particular, it tries to eliminate a lot of the overhead around remembering random incantions and one liners, allowing you to view the information you care about at a glance.

### Design goals
1. Single binary, zero config: "just works" out of the box
2. No root priveliges required for basic functionality, optional features with elevated privileges
3. Real-time, much like how tools like `htop` work. That said, it tries to be relatively lean and just do the networking aspect really well
4. Simple and intuitive navigation and ui 

### Who this is for
- Developers debugging local services (i.e. "what's taking up port 8443?") or monitoring servers
- Basically anyone who's tired of remembering `lsof -i :8080` vs `ss -tulnp` or wants a cleaner view for these

---

## Basic support


| Feature | Desc |
|---------|-------------|
| Connections view | Live TCP/UDP connections with process info |
| Ports view | Listening ports grouped by process |
| Interfaces view | Network interface stats (RX/TX bytes) |
| Tab navigation | Switch between views |
| Cursor navigation | Vim-style j/k movement with scrolling |
| Auto-refresh | 2-second polling interval |



## TODOs

- [ ] vim-style navigation, no mouse required

