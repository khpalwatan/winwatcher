package main

import (
	"context"

	"github.com/khpalwatan/winwatcher/internal/proc"
	"github.com/khpalwatan/winwatcher/internal/util"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the Wails application struct.
type App struct {
	ctx    context.Context
	cancel context.CancelFunc

	selectedPID int32
}

// NewApp creates the App.
func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx, a.cancel = context.WithCancel(ctx)
	go proc.Watch(a.ctx, a.getSelectedPID)
}

func (a *App) shutdown(ctx context.Context) {
	if w, h := runtime.WindowGetSize(ctx); w > 0 && h > 0 {
		util.SaveWindowState(w, h)
	}
	if a.cancel != nil {
		a.cancel()
	}
}

// GetTree returns the full process tree.
func (a *App) GetTree() []proc.Node {
	nodes, err := proc.BuildTree()
	if err != nil {
		return []proc.Node{}
	}
	return nodes
}

// SetSelectedPID records the user's selection for the watch loop.
func (a *App) SetSelectedPID(pid int32) {
	a.selectedPID = pid
}

// GetDetails returns details for one PID immediately.
func (a *App) GetDetails(pid int32) *proc.Details {
	return proc.GetDetails(pid)
}

// KillProcess terminates a process.
func (a *App) KillProcess(pid int32) proc.KillResult {
	return proc.KillProcess(pid)
}

// OpenURL opens the given URL in the user's default browser.
func (a *App) OpenURL(url string) {
	runtime.BrowserOpenURL(a.ctx, url)
}

// GetVersion returns the app version string.
func (a *App) GetVersion() string { return util.Version }

// GetAuthor returns the author display name.
func (a *App) GetAuthor() string { return util.Author }

// GetGitHubURL returns the author's GitHub URL.
func (a *App) GetGitHubURL() string { return util.GitHubURL }

func (a *App) getSelectedPID() int32 {
	return a.selectedPID
}