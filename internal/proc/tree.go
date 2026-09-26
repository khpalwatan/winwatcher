package proc

import (
	"fmt"
	"sort"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Node is one row in the process tree.
type Node struct {
	PID        int32   `json:"pid"`
	ParentPID  int32   `json:"parentPid"`
	Name       string  `json:"name"`
	CPUPercent float64 `json:"cpuPercent"`
	MemoryRSS  uint64  `json:"memoryRss"`
	Depth      int     `json:"depth"`
	IsSystem   bool    `json:"isSystem"`
	Tag        string  `json:"tag"`
}

var (
	ntdll                 = windows.NewLazySystemDLL("ntdll.dll")
	procNtQuerySystemInfo = ntdll.NewProc("NtQuerySystemInformation")
)

const (
	systemProcessInformation = 5
	statusInfoLengthMismatch = 0xC0000004
	initialBufSize           = 1 << 20
)

var (
	prevMu    sync.Mutex
	prevCPU   = map[int32]uint64{}
	prevTotal uint64
)

// BuildTree returns the full process list, ordered parent-first.
// Zombie processes (zero threads) are filtered out.
func BuildTree() ([]Node, error) {
	raw, totalCPU, err := queryProcesses()
	if err != nil {
		return nil, err
	}

	prevMu.Lock()
	oldTotal := prevTotal
	prevTotal = totalCPU
	prevMu.Unlock()

	totalDelta := float64(totalCPU - oldTotal)
	if totalDelta <= 0 {
		totalDelta = 1
	}

	// Filter zombies: a live process has at least one thread.
	live := raw[:0]
	for _, p := range raw {
		if p.threads > 0 {
			live = append(live, p)
		}
	}
	raw = live

	nodes := make([]Node, 0, len(raw))
	byPID := make(map[int32]int, len(raw))

	for _, p := range raw {
		prevMu.Lock()
		last, had := prevCPU[p.pid]
		prevCPU[p.pid] = p.cpuTime
		prevMu.Unlock()

		var cpuPct float64
		if had && p.cpuTime > last {
			cpuPct = float64(p.cpuTime-last) / totalDelta * 100
		}

		tag := tagForPID(p.pid)

		byPID[p.pid] = len(nodes)
		nodes = append(nodes, Node{
			PID:        p.pid,
			ParentPID:  p.parentPID,
			Name:       displayName(p.pid, p.name),
			CPUPercent: cpuPct,
			MemoryRSS:  p.memory,
			IsSystem:   tag == TagSystem,
			Tag:        tag,
		})
	}

	children := make(map[int32][]int32, len(raw))
	for _, p := range raw {
		if p.pid == 0 {
			continue
		}
		children[p.parentPID] = append(children[p.parentPID], p.pid)
	}
	for k := range children {
		sort.Slice(children[k], func(i, j int) bool {
			return children[k][i] < children[k][j]
		})
	}

	out := make([]Node, 0, len(nodes))
	visited := make(map[int32]bool, len(nodes))

	var walk func(pid int32, depth int)
	walk = func(pid int32, depth int) {
		if visited[pid] {
			return
		}
		idx, ok := byPID[pid]
		if !ok {
			return
		}
		visited[pid] = true
		n := nodes[idx]
		n.Depth = depth
		out = append(out, n)
		for _, c := range children[pid] {
			walk(c, depth+1)
		}
	}

	if _, ok := byPID[0]; ok {
		walk(0, 0)
	}
	for _, p := range raw {
		if !visited[p.pid] {
			walk(p.pid, 0)
		}
	}

	return out, nil
}

type rawProc struct {
	pid       int32
	parentPID int32
	name      string
	cpuTime   uint64
	memory    uint64
	threads   uint32
}

func queryProcesses() ([]rawProc, uint64, error) {
	bufSize := uint32(initialBufSize)
	var buf []byte

	for {
		buf = make([]byte, bufSize)
		var retLen uint32
		r, _, _ := procNtQuerySystemInfo.Call(
			uintptr(systemProcessInformation),
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(bufSize),
			uintptr(unsafe.Pointer(&retLen)),
		)
		if r == statusInfoLengthMismatch {
			bufSize = retLen + 64*1024
			continue
		}
		if r != 0 {
			return nil, 0, fmt.Errorf("NtQuerySystemInformation failed: 0x%x", r)
		}
		buf = buf[:retLen]
		break
	}

	return parseProcesses(buf)
}

const (
	offUniquePID       = 0x50
	offInheritedFromID = 0x58
	offWorkingSet      = 0x90
	offKernelTime      = 0xA8
	offUserTime        = 0xB0
	offImageName       = 0x38
	offNumThreads      = 0x04
)

func parseProcesses(buf []byte) ([]rawProc, uint64, error) {
	out := make([]rawProc, 0, 512)
	var totalCPU uint64

	off := uintptr(0)
	base := uintptr(unsafe.Pointer(&buf[0]))
	bufLen := uintptr(len(buf))

	for {
		if off+0x100 > bufLen {
			break
		}
		p := unsafe.Pointer(base + off)

		next := *(*uint32)(p)
		threads := *(*uint32)(unsafe.Pointer(base + off + offNumThreads))
		pid := *(*int32)(unsafe.Pointer(base + off + offUniquePID))
		ppid := *(*int32)(unsafe.Pointer(base + off + offInheritedFromID))
		mem := *(*uint64)(unsafe.Pointer(base + off + offWorkingSet))
		kt := *(*uint64)(unsafe.Pointer(base + off + offKernelTime))
		ut := *(*uint64)(unsafe.Pointer(base + off + offUserTime))
		cpu := kt + ut

		nameLen := *(*uint16)(unsafe.Pointer(base + off + offImageName))
		namePtr := *(*uintptr)(unsafe.Pointer(base + off + offImageName + 8))
		name := ""
		if nameLen > 0 && namePtr != 0 {
			u16 := unsafe.Slice((*uint16)(unsafe.Pointer(namePtr)), nameLen/2)
			name = windows.UTF16ToString(u16)
		}

		out = append(out, rawProc{
			pid:       pid,
			parentPID: ppid,
			name:      name,
			cpuTime:   cpu,
			memory:    mem,
			threads:   threads,
		})
		totalCPU += cpu

		if next == 0 {
			break
		}
		off += uintptr(next)
	}

	return out, totalCPU, nil
}

func displayName(pid int32, name string) string {
	if name != "" {
		return name
	}
	switch pid {
	case 0:
		return "System Idle Process"
	case 4:
		return "System"
	}
	return fmt.Sprintf("PID %d", pid)
}

func isSystemPID(pid int32) bool {
	return pid == 0 || pid == 4
}