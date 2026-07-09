# riptide

**Measure and watch your internet connection — from the terminal.**

A polished Go TUI with a startup menu, one-shot speed tests, and a live bandwidth monitor. Centered cards, smooth graphs, no config required.

[![terminal](https://img.shields.io/badge/terminal-TUI-39d0d8?style=flat-square)](https://github.com/Foxemsx/riptide)
[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev)
[![Linux](https://img.shields.io/badge/Linux-supported-2ea44f?style=flat-square&logo=linux&logoColor=white)](https://github.com/Foxemsx/riptide)
[![Windows](https://img.shields.io/badge/Windows-supported-0078D6?style=flat-square&logo=windows&logoColor=white)](https://github.com/Foxemsx/riptide)
[![License](https://img.shields.io/badge/license-MIT-blue?style=flat-square)](LICENSE)

<p align="center">
  <img src="assets/showcase.gif?v=3" alt="riptide demo" width="720">
</p>

---

## Modes

| | |
|:---|:---|
| **Speed Test** | One-shot download, upload, and ping. Parallel connections, peak rates, timed phases. |
| **Bandwidth Monitor** | Live view of *real* PC traffic (OS counters only — no test load). Peaks, uptime, pause. |

Both modes share the same card UI, accent colors (teal ↓ / amber ↑), and keyboard controls.

---

## Screenshots

<p align="center">
  <img src="assets/mainmenu.png?v=3" alt="Main menu" width="48%">
  &nbsp;
  <img src="assets/home.png?v=3" alt="Speed test" width="48%">
</p>
<p align="center">
  <sub><b>Main menu</b> · pick Speed Test, Bandwidth, or Exit &nbsp;&nbsp;|&nbsp;&nbsp; <b>Speed Test</b> · live graphs mid-run</sub>
</p>

<p align="center">
  <img src="assets/finished.png?v=3" alt="Finished summary" width="48%">
  &nbsp;
  <img src="assets/bandwidth.png?v=3" alt="Bandwidth monitor" width="48%">
</p>
<p align="center">
  <sub><b>Finished</b> · peaks + ping summary &nbsp;&nbsp;|&nbsp;&nbsp; <b>Bandwidth</b> · live DL/UL of your real connection</sub>
</p>

<p align="center">
  <img src="assets/helpmenu.png?v=3" alt="Help overlay" width="42%">
</p>
<p align="center">
  <sub><b>Help</b> · press <code>?</code> anytime for controls · <code>esc</code> / <code>m</code> back to menu</sub>
</p>

---

## Features

- **Startup menu** — card buttons with hotkeys `1` / `2` / `3`, keyboard or mouse
- **High-res graphs** — eighth-block bars, fire gradients, peak spark, age fade
- **Smooth numbers** — lerped display values instead of hard snaps
- **Units on the fly** — `c` cycles Mbps · KB/s · MB/s · GB/s
- **Compact mode** — `t` hides the large logo when space is tight
- **Clean chrome** — VS-style `#191a1b` canvas, rounded cards, accent chips
- **Graceful errors** — no stack traces when the network is down

---

## Quick start

**Linux** (installer — no sudo):

```sh
curl -fsSL https://raw.githubusercontent.com/Foxemsx/riptide/main/install.sh | sh
riptide
```

**Anywhere with Go 1.23+**:

```sh
go install github.com/Foxemsx/riptide/cmd/riptide@latest
riptide
```

**From source**:

```sh
git clone https://github.com/Foxemsx/riptide
cd riptide
go build -o riptide ./cmd/riptide    # Windows: go build -o riptide.exe ./cmd/riptide
./riptide
```

> Put `$(go env GOPATH)/bin` on your `PATH` if `riptide` is not found after `go install`.

Uninstall (Linux):

```sh
curl -fsSL https://raw.githubusercontent.com/Foxemsx/riptide/main/uninstall.sh | sh
```

---

## Usage

```sh
riptide              # main menu → Speed Test or Bandwidth
riptide --compact    # skip the large logo
```

| Flag | Default | Description |
|------|---------|-------------|
| `--compact` | `false` | Tagline only (no large logo) |
| `--theme` | `default` | Reserved for future palettes |

### Controls

| Key | Action |
|-----|--------|
| `←` `→` / `j` `k` | Move in the menu |
| `1` `2` `3` | Jump to a menu option |
| `enter` | Select |
| `c` | Cycle units |
| `r` | Restart test / monitor |
| `p` | Pause / resume (**Bandwidth** only) |
| `t` | Toggle compact logo |
| `?` | Help overlay |
| `esc` / `m` | Back to main menu |
| `q` / `ctrl+c` | Quit |

---

## License

[MIT](LICENSE) — free to use, modify, and redistribute with the license notice.
