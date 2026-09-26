package proc

import (
	"context"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// RefreshInterval is how often the watch loop emits updates.
const RefreshInterval = 2 * time.Second

// GraphHistory is the number of samples kept for the CPU/memory graph.
// At 2s intervals, 60 samples = 2 minutes of history.
const GraphHistory = 60

// FileHandle is one open file held by a process.
type FileHandle struct {
	Path string `json:"path"`
}

// NetConn is one network connection owned by a process.
type NetConn struct {
	Protocol   string `json:"protocol"`
	LocalAddr  string `json:"localAddr"`
	RemoteAddr string `json:"remoteAddr"`
	State      string `json:"state"`
}

// Details is the data shown in the right pane.
type Details struct {
	PID         int32        `json:"pid"`
	CommandLine string       `json:"commandLine"`
	Threads     int32        `json:"threads"`
	Files       []FileHandle `json:"files"`
	FilesErr    string       `json:"filesErr,omitempty"`
	Net         []NetConn    `json:"net"`
}

// Sample is one point on the CPU/memory graph.
type Sample struct {
	CPU    float64 `json:"cpu"`    // 0-100
	Memory uint64  `json:"memory"` // bytes
}

// graphBuf is a ring buffer of samples, one per tick.
type graphBuf struct {
	mu      sync.Mutex
	samples []Sample
	head    int
	full    bool
}

var graph = &graphBuf{samples: make([]Sample, GraphHistory)}

func (g *graphBuf) push(s Sample) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.samples[g.head] = s
	g.head = (g.head + 1) % len(g.samples)
	if g.head == 0 {
		g.full = true
	}
}

func (g *graphBuf) snapshot() []Sample {
	g.mu.Lock()
	defer g.mu.Unlock()
	n := g.head
	if g.full {
		n = len(g.samples)
	}
	out := make([]Sample, 0, n)
	start := 0
	if g.full {
		start = g.head
	}
	for i := 0; i < n; i++ {
		out = append(out, g.samples[(start+i)%len(g.samples)])
	}
	return out
}

// GetGraph returns the current sample history for the frontend graph.
func GetGraph() []Sample {
	return graph.snapshot()
}

// GetDetails gathers command line, threads, open files, and network
// connections for one PID.
func GetDetails(pid int32) *Details {
	d := &Details{
		PID:   pid,
		Files: []FileHandle{},
		Net:   []NetConn{},
	}

	p, err := process.NewProcess(pid)
	if err != nil {
		return d
	}

	if cmd, err := p.Cmdline(); err == nil {
		d.CommandLine = cmd
	}
	if n, err := p.NumThreads(); err == nil {
		d.Threads = n
	}

	files, err := p.OpenFiles()
	if err != nil {
		d.FilesErr = "Access denied. This process is protected. Run winwatcher as administrator for full access."
	} else {
		for _, f := range files {
			d.Files = append(d.Files, FileHandle{Path: f.Path})
		}
	}

	d.Net = networkForPID(pid)
	return d
}

func networkForPID(pid int32) []NetConn {
	conns, err := net.Connections("all")
	if err != nil {
		return []NetConn{}
	}
	out := make([]NetConn, 0, 4)
	for _, c := range conns {
		if c.Pid != pid {
			continue
		}
		out = append(out, NetConn{
			Protocol:   protoName(c.Type),
			LocalAddr:  fmtAddr(c.Laddr.IP, c.Laddr.Port),
			RemoteAddr: fmtAddr(c.Raddr.IP, c.Raddr.Port),
			State:      c.Status,
		})
	}
	return out
}

func protoName(t uint32) string {
	switch t {
	case 1:
		return "tcp4"
	case 2:
		return "tcp6"
	case 3:
		return "udp4"
	case 4:
		return "udp6"
	}
	return "sock"
}

func fmtAddr(ip string, port uint32) string {
	if ip == "" || ip == "0.0.0.0" || ip == "::" {
		return "*:" + itoa(port)
	}
	return ip + ":" + itoa(port)
}

func itoa(n uint32) string {
	if n == 0 {
		return "0"
	}
	var b [10]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

// Watch runs in a goroutine, emits events to the frontend every tick.
func Watch(ctx context.Context, selectedPID func() int32) {
	ticker := time.NewTicker(RefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			nodes, err := BuildTree()
			if err != nil {
				continue
			}
			runtime.EventsEmit(ctx, "proc:tree", nodes)

			// Aggregate totals for the graph.
			var totalCPU float64
			var totalMem uint64
			for _, n := range nodes {
				totalCPU += n.CPUPercent
				totalMem += n.MemoryRSS
			}
			graph.push(Sample{CPU: totalCPU, Memory: totalMem})
			runtime.EventsEmit(ctx, "graph:update", graph.snapshot())

			pid := selectedPID()
			if pid > 0 {
				runtime.EventsEmit(ctx, "proc:details", GetDetails(pid))
			}
		}
	}
}