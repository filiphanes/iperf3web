package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

// TestParams holds all configurable iperf3 options.
type TestParams struct {
	Server         string `json:"server"`
	Port           int    `json:"port"`
	Protocol       string `json:"protocol"`        // "tcp" | "udp"
	Direction      string `json:"direction"`       // "upload" | "download" | "bidir"
	Duration       int    `json:"duration"`
	Parallel       int    `json:"parallel"`
	Bandwidth      string `json:"bandwidth"`       // e.g. "100M", "" for unlimited/default
	Window         string `json:"window"`          // e.g. "256K"
	Length         string `json:"length"`          // e.g. "128K" socket buffer/payload length
	MSS            int    `json:"mss"`
	NoDelay        bool   `json:"no_delay"`
	IPv4           bool   `json:"ipv4"`            // force IPv4 (-4)
	IPv6           bool   `json:"ipv6"`            // force IPv6 (-6)
	OmitSecs       int    `json:"omit_secs"`
	ZeroCopy       bool   `json:"zero_copy"`       // use zero-copy send (-Z)
	ConnectTimeout int    `json:"connect_timeout"` // connection setup timeout in ms (0=default)
}

// SSEMsg is sent to all SSE subscribers.
type SSEMsg struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// testDone carries the final result back to the caller.
type testDone struct {
	params    TestParams
	end       *EndData
	intervals []float64 // bps per non-omitted interval
	errStr    string
}

// Runner manages one iperf3 subprocess and fans SSE messages to all clients.
type Runner struct {
	mu       sync.Mutex
	cancel   context.CancelFunc
	running  bool
	params   *TestParams
	clientMu sync.RWMutex
	clients  map[chan SSEMsg]struct{}
}

func newRunner() *Runner {
	return &Runner{clients: make(map[chan SSEMsg]struct{})}
}

func (r *Runner) subscribe() chan SSEMsg {
	ch := make(chan SSEMsg, 128)
	r.clientMu.Lock()
	r.clients[ch] = struct{}{}
	r.clientMu.Unlock()
	return ch
}

func (r *Runner) unsubscribe(ch chan SSEMsg) {
	r.clientMu.Lock()
	delete(r.clients, ch)
	r.clientMu.Unlock()
}

func (r *Runner) broadcast(msg SSEMsg) {
	r.clientMu.RLock()
	defer r.clientMu.RUnlock()
	for ch := range r.clients {
		select {
		case ch <- msg:
		default: // drop slow clients
		}
	}
}

func (r *Runner) isRunning() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.running
}

func (r *Runner) getParams() *TestParams {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.params
}

func (r *Runner) start(params TestParams, onDone func(testDone)) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.running {
		return fmt.Errorf("test already running")
	}
	r.running = true
	r.params = &params
	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel
	go r.run(ctx, params, onDone)
	return nil
}

func (r *Runner) stop() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cancel != nil {
		r.cancel()
	}
}

func buildArgs(p TestParams) []string {
	port := p.Port
	if port == 0 {
		port = 5201
	}
	args := []string{
		"-c", p.Server,
		"--json-stream",
		"--forceflush",
		"-p", strconv.Itoa(port),
		"-t", strconv.Itoa(p.Duration),
		"-i", "1",
	}
	if p.Protocol == "udp" {
		args = append(args, "-u")
		bw := p.Bandwidth
		if bw == "" {
			bw = "10M" // iperf3 defaults to 1M for UDP; 10M is more useful
		}
		args = append(args, "-b", bw)
	} else if p.Bandwidth != "" {
		args = append(args, "-b", p.Bandwidth)
	}
	if p.Direction == "download" {
		args = append(args, "-R")
	} else if p.Direction == "bidir" {
		args = append(args, "--bidir")
	}
	if p.Parallel > 1 {
		args = append(args, "-P", strconv.Itoa(p.Parallel))
	}
	if p.Window != "" {
		args = append(args, "-w", p.Window)
	}
	if p.Length != "" {
		args = append(args, "-l", p.Length)
	}
	if p.MSS > 0 {
		args = append(args, "-M", strconv.Itoa(p.MSS))
	}
	if p.NoDelay {
		args = append(args, "-N")
	}
	if p.IPv4 {
		args = append(args, "-4")
	} else if p.IPv6 {
		args = append(args, "-6")
	}
	if p.OmitSecs > 0 {
		args = append(args, "-O", strconv.Itoa(p.OmitSecs))
	}
	if p.ZeroCopy {
		args = append(args, "-Z")
	}
	if p.ConnectTimeout > 0 {
		args = append(args, "--connect-timeout", strconv.Itoa(p.ConnectTimeout))
	}
	return args
}

// cleanStderr turns raw iperf3 stderr into a short, user-friendly message.
func cleanStderr(s string) string {
	s = strings.TrimSpace(s)
	// Strip common "iperf3: error - " prefix
	s = strings.TrimPrefix(s, "iperf3: error - ")
	// Map known patterns to friendlier text
	sl := strings.ToLower(s)
	switch {
	case strings.Contains(sl, "connection refused"):
		return "Connection refused — is iperf3 running on the server?"
	case strings.Contains(sl, "connection timed out") || strings.Contains(sl, "timed out"):
		return "Connection timed out — server unreachable or firewall blocking port"
	case strings.Contains(sl, "no route to host"):
		return "No route to host — check server address"
	case strings.Contains(sl, "name or service not known") || strings.Contains(sl, "nodename nor servname"):
		return "Hostname not found — check server address"
	case strings.Contains(sl, "busy"):
		return "Server is busy running another test — try again later"
	}
	return s
}

func (r *Runner) run(ctx context.Context, params TestParams, onDone func(testDone)) {
	defer func() {
		r.mu.Lock()
		r.running = false
		r.params = nil
		if r.cancel != nil {
			r.cancel()
			r.cancel = nil
		}
		r.mu.Unlock()
	}()

	exe := iperf3Executable()
	args := buildArgs(params)
	log.Printf("Running: iperf3 %s", strings.Join(args, " "))

	cmd := exec.CommandContext(ctx, exe, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		r.broadcast(SSEMsg{Type: "error", Payload: err.Error()})
		r.broadcast(SSEMsg{Type: "done", Payload: nil})
		onDone(testDone{params: params, errStr: err.Error()})
		return
	}
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		r.broadcast(SSEMsg{Type: "error", Payload: err.Error()})
		r.broadcast(SSEMsg{Type: "done", Payload: nil})
		onDone(testDone{params: params, errStr: err.Error()})
		return
	}

	// Capture stderr (iperf3 writes errors there before any JSON)
	errCh := make(chan string, 1)
	go func() {
		scanner := bufio.NewScanner(stderr)
		var lines []string
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		errCh <- strings.Join(lines, "\n")
	}()

	var intervals []float64
	var endData *EndData
	var jsonErrMsg string

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 1<<20), 1<<20) // 1 MB for large end events

	for scanner.Scan() {
		raw := scanner.Bytes()
		var event StreamEvent
		if err := json.Unmarshal(raw, &event); err != nil {
			continue
		}
		switch event.Event {
		case "start":
			var d StartData
			json.Unmarshal(event.Data, &d)
			r.broadcast(SSEMsg{Type: "start", Payload: d})

		case "interval":
			var d IntervalData
			json.Unmarshal(event.Data, &d)
			if !d.Sum.Omitted {
				intervals = append(intervals, d.Sum.BitsPerSecond)
			}
			r.broadcast(SSEMsg{Type: "interval", Payload: d})

		case "end":
			var d EndData
			json.Unmarshal(event.Data, &d)
			endData = &d
			r.broadcast(SSEMsg{Type: "end", Payload: d})

		case "error":
			var msg string
			json.Unmarshal(event.Data, &msg)
			jsonErrMsg = msg
			r.broadcast(SSEMsg{Type: "error", Payload: msg})
		}
	}

	cmd.Wait()
	stderrMsg := <-errCh

	done := testDone{params: params, end: endData, intervals: intervals}
	if endData == nil {
		switch {
		case jsonErrMsg != "":
			// Already broadcast above; just record it.
			done.errStr = jsonErrMsg
		case stderrMsg != "":
			done.errStr = cleanStderr(stderrMsg)
			r.broadcast(SSEMsg{Type: "error", Payload: done.errStr})
		default:
			done.errStr = "test ended with no results (was it cancelled?)"
		}
	}

	r.broadcast(SSEMsg{Type: "done", Payload: nil})
	onDone(done)
}
