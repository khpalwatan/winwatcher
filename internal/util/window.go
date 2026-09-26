package util

import (
	"encoding/json"
	"os"
	"path/filepath"
	"syscall"
)

// WindowState holds the persisted window size.
type WindowState struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

const (
	defaultWidth  = 1100
	defaultHeight = 700
	minWidth      = 720
	minHeight     = 480

	marginX = 60
	marginY = 90
)

var (
	user32           = syscall.NewLazyDLL("user32.dll")
	getSystemMetrics = user32.NewProc("GetSystemMetrics")
)

const (
	SM_CXSCREEN = 0
	SM_CYSCREEN = 1
)

func stateFilePath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "winwatcher", "window.json")
}

// screenSize returns the primary display's pixel size using GetSystemMetrics.
// This is simpler than GetDeviceCaps and works on all Windows versions.
func screenSize() (int, int) {
	w, _, _ := getSystemMetrics.Call(SM_CXSCREEN)
	h, _, _ := getSystemMetrics.Call(SM_CYSCREEN)
	return int(w), int(h)
}

// LoadWindowState returns the saved window size if valid, otherwise a
// default that fits comfortably on the current screen with margins.
func LoadWindowState() WindowState {
	sw, sh := screenSize()

	maxW := defaultWidth
	maxH := defaultHeight
	if sw > 0 {
		if v := sw - marginX; v > minWidth && v < maxW {
			maxW = v
		}
	}
	if sh > 0 {
		if v := sh - marginY; v > minHeight && v < maxH {
			maxH = v
		}
	}

	s := WindowState{Width: maxW, Height: maxH}

	if path := stateFilePath(); path != "" {
		if data, err := os.ReadFile(path); err == nil {
			var loaded WindowState
			if err := json.Unmarshal(data, &loaded); err == nil {
				if loaded.Width >= minWidth {
					s.Width = loaded.Width
				}
				if loaded.Height >= minHeight {
					s.Height = loaded.Height
				}
			}
		}
	}

	if sw > 0 && s.Width > sw-marginX {
		s.Width = sw - marginX
	}
	if sh > 0 && s.Height > sh-marginY {
		s.Height = sh - marginY
	}
	if s.Width < minWidth {
		s.Width = minWidth
	}
	if s.Height < minHeight {
		s.Height = minHeight
	}

	return s
}

// SaveWindowState persists the given size.
func SaveWindowState(w, h int) {
	if w < minWidth || h < minHeight {
		return
	}
	path := stateFilePath()
	if path == "" {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	data, err := json.Marshal(WindowState{Width: w, Height: h})
	if err != nil {
		return
	}
	_ = os.WriteFile(path, data, 0o644)
}