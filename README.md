<div align="center">

<img src="https://github.com/user-attachments/assets/f53a0b15-c44d-4267-ade0-4dc8515363e3" alt="winwatcher" width="160" height="160" />

# winwatcher

**Live process tree with real-time file and network activity for Windows.**

[![Release](https://img.shields.io/github/v/release/khpalwatan/winwatcher?style=flat-square&color=22d3ee)](https://github.com/khpalwatan/winwatcher/releases)
[![License](https://img.shields.io/github/license/khpalwatan/winwatcher?style=flat-square&color=22d3ee)](https://github.com/khpalwatan/winwatcher/blob/main/LICENSE)
[![Platform](https://img.shields.io/badge/platform-Windows%2010%20%7C%2011-0078D4?style=flat-square)](#)
[![Made with Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev)
[![Made with Wails](https://img.shields.io/badge/Wails-v2.12-DF0000?style=flat-square)](https://wails.io)

</div>

---

## What it is

winwatcher shows the full Windows process tree and, for any selected process, live-updates the files and network connections it's currently touching.

**The killer feature:** when a file is locked by a mystery process, open winwatcher, search for the filename, and instantly see which process has it open.

Think Process Explorer — but modern, beautiful, keyboard-friendly, and fast.

---

## Features

| | |
|---|---|
| **Full process tree** | Parent-child relationships, PID, CPU %, memory, SYSTEM / SERVICE / USER tags |
| **Open file handles** | Per process — see exactly which files a process is holding |
| **Live network connections** | Per process — local:port → remote:port, state, protocol |
| **Live graphs** | CPU and memory scrolling at the bottom, updated every 2 seconds |
| **Three themes** | Dark, AMOLED Black (pure black), White Glass (frosted) |
| **Alert highlighting** | Processes using >25% CPU or >500 MB RAM are flagged red |
| **Kill Process** | One click with a Windows-native confirmation dialog |
| **Filter box** | Find any process by name or PID instantly |
| **Keyboard-first** | `F11` fullscreen · `/` focus filter · `T` cycle theme · `Esc` exit fullscreen |
| **Single .exe** | No installer, no dependencies — uses the WebView2 runtime already on Windows |

---

## Install

1. Go to [**Releases**](https://github.com/khpalwatan/winwatcher/releases)
2. Download the latest `winwatcher-windows-amd64.exe`
3. Double-click it. That's it.

No installer. No admin rights. No dependencies to install.

---

## Screenshots

### Dark theme

<img src="https://github.com/user-attachments/assets/b1b38d9e-711e-4d11-8996-a5807e884b47" alt="winwatcher dark theme" width="100%" />

### White Glass theme

<img src="https://github.com/user-attachments/assets/d3e10c9d-9378-4f2d-8616-1e619ae89b67" alt="winwatcher white glass theme" width="100%" />

### AMOLED Black theme

<img src="https://github.com/user-attachments/assets/63e863a8-5d59-4c22-ba62-ee12d921788c" alt="winwatcher amoled theme" width="100%" />

---

## Keyboard shortcuts

| Key | Action |
|---|---|
| `F11` | Toggle fullscreen |
| `/` | Focus the filter box |
| `Esc` | Clear filter / exit fullscreen |
| `T` | Cycle theme (Dark → AMOLED → White Glass) |

---

## Building from source

**Requirements:**

- [Go](https://go.dev) 1.23+
- [Node.js](https://nodejs.org) 18+
- [Wails CLI](https://wails.io) v2.12+

**Steps:**

```bash
# Install the Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0

# Clone the repo
git clone https://github.com/khpalwatan/winwatcher
cd winwatcher

# Install frontend dependencies
cd frontend
npm install
cd ..

# Run in dev mode (hot reload)
wails dev

# Or build the release binary
wails build -platform windows/amd64 -clean
```

The output binary is written to `build/bin/winwatcher.exe`.

---

## Tech stack

- **[Wails v2](https://wails.io)** — Go + WebView2 desktop framework
- **[Go 1.23+](https://go.dev)** — backend
- **Vanilla HTML / CSS / JavaScript** — frontend, no framework
- **[gopsutil](https://github.com/shirou/gopsutil)** — process and network info
- **Direct `NtQuerySystemInformation` syscall** — the fast Windows kernel API Task Manager uses, called from Go with no C compiler

---

## Known limitations

- System processes (`svchost.exe`, `csrss.exe`, etc.) show **"Access denied"** for open files unless winwatcher is run as administrator. This is a Windows security boundary — Task Manager has the same limitation.
- **Windows 10 / 11 only.**
- Alpha release — expect rough edges. Report issues on the [Issues](https://github.com/khpalwatan/winwatcher/issues) page.

---

## Related

- **[portswarden](https://github.com/khpalwatan/portswarden)** — Find and free Windows ports. Kill EADDRINUSE in one command.

---

## License

[MIT](LICENSE) © 2026 [WAseem KHan](https://github.com/khpalwatan)

<div align="center">

**Built by [WAseem KHan](https://github.com/khpalwatan)**

⭐ If winwatcher is useful to you, give it a star — it helps a lot.

---

<sub>Crafted with care by <a href="https://github.com/khpalwatan">WAseem KHan</a> · Idea, design, and code — all his.</sub>

</div>
