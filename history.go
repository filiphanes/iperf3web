package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// HistoryEntry stores one completed test result.
type HistoryEntry struct {
	ID          string     `json:"id"`
	Timestamp   time.Time  `json:"timestamp"`
	Params      TestParams `json:"params"`
	SendBps     float64    `json:"send_bps"`
	RecvBps     float64    `json:"recv_bps"`
	JitterMs    float64    `json:"jitter_ms,omitempty"`
	LostPercent float64    `json:"lost_percent,omitempty"`
	Intervals   []float64  `json:"intervals"` // bps per second
	DurationS   float64    `json:"duration_s"`
	Error       string     `json:"error,omitempty"`
}

// HistoryStore is a thread-safe, file-backed list of test results.
type HistoryStore struct {
	mu      sync.RWMutex
	path    string
	entries []HistoryEntry
}

func newHistoryStore(path string) *HistoryStore {
	s := &HistoryStore{path: path}
	s.load()
	return s
}

func (s *HistoryStore) load() {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	json.Unmarshal(data, &s.entries)
}

func (s *HistoryStore) save() {
	s.mu.RLock()
	data, err := json.MarshalIndent(s.entries, "", "  ")
	s.mu.RUnlock()
	if err == nil {
		os.WriteFile(s.path, data, 0o644)
	}
}

func (s *HistoryStore) add(entry HistoryEntry) {
	s.mu.Lock()
	s.entries = append([]HistoryEntry{entry}, s.entries...)
	if len(s.entries) > 500 {
		s.entries = s.entries[:500]
	}
	s.mu.Unlock()
	s.save()
}

func (s *HistoryStore) getAll() []HistoryEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]HistoryEntry, len(s.entries))
	copy(out, s.entries)
	return out
}

func (s *HistoryStore) delete(id string) {
	s.mu.Lock()
	for i, e := range s.entries {
		if e.ID == id {
			s.entries = append(s.entries[:i], s.entries[i+1:]...)
			break
		}
	}
	s.mu.Unlock()
	s.save()
}

func (s *HistoryStore) clear() {
	s.mu.Lock()
	s.entries = nil
	s.mu.Unlock()
	s.save()
}

// newEntry builds a HistoryEntry from a completed test.
func newEntry(done testDone) HistoryEntry {
	entry := HistoryEntry{
		ID:        fmt.Sprintf("%d", time.Now().UnixMilli()),
		Timestamp: time.Now(),
		Params:    done.params,
		Intervals: done.intervals,
		Error:     done.errStr,
	}
	if done.end != nil {
		entry.SendBps = done.end.SumSent.BitsPerSecond
		entry.RecvBps = done.end.SumReceived.BitsPerSecond
		entry.DurationS = done.end.SumSent.End
		if done.params.Protocol == "udp" {
			entry.JitterMs = done.end.SumReceived.JitterMs
			entry.LostPercent = done.end.SumReceived.LostPercent
		}
	}
	return entry
}
