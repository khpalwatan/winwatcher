package proc

import (
	"fmt"

	"github.com/shirou/gopsutil/v3/process"
)

// KillResult reports the outcome of a KillProcess call.
type KillResult struct {
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
}

// KillProcess terminates the process with the given PID.
//
// Returns a KillResult rather than an error so the frontend can display
// a clean message instead of a Go error string.
//
// Note on privileges: some processes require elevation. If the caller
// does not have rights, the message explains that clearly.
func KillProcess(pid int32) KillResult {
	if pid <= 4 {
		return KillResult{
			OK:      false,
			Message: "Cannot kill system processes (PID 0-4).",
		}
	}

	p, err := process.NewProcess(pid)
	if err != nil {
		return KillResult{
			OK:      false,
			Message: "Process not found. It may have already exited.",
		}
	}

	if err := p.Kill(); err != nil {
		return KillResult{
			OK: false,
			Message: fmt.Sprintf(
				"Could not kill process. It may be protected or you may need to run winwatcher as administrator. (%v)",
				err,
			),
		}
	}

	return KillResult{OK: true}
}