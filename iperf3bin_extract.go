//go:build release

package main

import (
	"log"
	"os"
	"sync"
)

var (
	iperf3Once      sync.Once
	iperf3CachedPath string
)

// extractIperf3 writes data to a temp file, marks it executable, and returns
// its path. The result is cached so extraction only happens once per process.
// Falls back to "iperf3" (PATH) if extraction fails.
func extractIperf3(data []byte, ext string) string {
	iperf3Once.Do(func() {
		f, err := os.CreateTemp("", "iperf3-*"+ext)
		if err != nil {
			log.Printf("warning: failed to create temp file for embedded iperf3: %v", err)
			return
		}
		if _, err := f.Write(data); err != nil {
			f.Close()
			os.Remove(f.Name())
			log.Printf("warning: failed to write embedded iperf3: %v", err)
			return
		}
		f.Close()
		if err := os.Chmod(f.Name(), 0755); err != nil {
			os.Remove(f.Name())
			log.Printf("warning: failed to chmod embedded iperf3: %v", err)
			return
		}
		iperf3CachedPath = f.Name()
	})
	if iperf3CachedPath == "" {
		return "iperf3" // fall back to PATH
	}
	return iperf3CachedPath
}
