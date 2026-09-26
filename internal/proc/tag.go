package proc

import (
	"golang.org/x/sys/windows"
)

// Tag classifies a process as system, service, or user.
// Used for the colored badge on each row.
const (
	TagSystem  = "system"
	TagService = "service"
	TagUser    = "user"
)

// tagCache avoids re-querying the same PID every refresh.
// A process's account never changes while it's running.
var tagCache = map[int32]string{}

// tagForPID returns "system", "service", or "user".
func tagForPID(pid int32) string {
	if v, ok := tagCache[pid]; ok {
		return v
	}

	tag := classify(pid)
	tagCache[pid] = tag
	return tag
}

func classify(pid int32) string {
	// PID 0 and 4 are the kernel's own processes.
	if pid == 0 || pid == 4 {
		return TagSystem
	}

	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		// Cannot open -- protected. Treat as system.
		return TagSystem
	}
	defer windows.CloseHandle(h)

	var token windows.Token
	if err := windows.OpenProcessToken(h, windows.TOKEN_QUERY, &token); err != nil {
		return TagSystem
	}
	defer token.Close()

	user, err := token.GetTokenUser()
	if err != nil {
		return TagSystem
	}

	sid := user.User.Sid.String()
	switch sid {
	case "S-1-5-18": // LocalSystem
		return TagSystem
	case "S-1-5-19": // LocalService
		return TagService
	case "S-1-5-20": // NetworkService
		return TagService
	}
	return TagUser
}