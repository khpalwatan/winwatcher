<h1 align="center">winwatcher</h1>

<p align="center">
  <strong>Live process tree with real-time file and network activity for Windows.</strong>
</p>

<p align="center">
  <a href="https://github.com/khpalwatan/Win-Watcher/releases">
    <img src="https://img.shields.io/github/v/release/khpalwatan/Win-Watcher?style=flat-square" alt="Release" />
  </a>
  <a href="https://github.com/khpalwatan/Win-Watcher/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/khpalwatan/Win-Watcher?style=flat-square" alt="License" />
  </a>
  <img src="https://img.shields.io/badge/platform-Windows%2010%2F11-blue?style=flat-square" alt="Platform" />
</p>

---

## What it is

winwatcher shows the full Windows process tree, and for any selected process, live-updates the files and network connections it's currently touching.

**The killer feature:** when a file is locked by a mystery process, you open winwatcher, search for the filename, and immediately see which process has it open.

Think Process Explorer — but modern, beautiful, keyboard-friendly, and fast.

## Features

- **Full process tree** with parent-child relationships, PID, CPU %, memory
- **Open file handles** per process
- **Live network connections** per process — local:port to remote:port, state, protocol
- **Live CPU + memory graphs** at the bottom, updating every 2 seconds
- **Three themes** — Dark, AMOLED Black, White Glass
- **Alert highlighting** — processes using >25% CPU or >500 MB RAM are marked red
- **Kill Process** button with confirmation
- **Filter box** to find any process by name or PID
- **Keyboard shortcuts:** `F11` fullscreen, `/` focus filter, `T` cycle theme, `Esc` exit fullscreen
- **Single .exe** — no installation, no dependencies (uses the WebView2 runtime already on Windows)

## Install

1. Go to [Releases](https://github.com/khpalwatan/Win-Watcher/releases)
2. Download the latest `winwatcher-windows-amd64.exe`
3. Run it. That's it.

No installer, no admin rights, no dependencies to install.

## Screenshots

Coming soon.

## Building from source

Requirements:
- Go 1.23+
- Node.js 18+
- Wails CLI v2.12+

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0
git clone https://github.com/khpalwatan/Win-Watcher
cd Win-Watcher
cd frontend && npm install && cd ..
wails dev