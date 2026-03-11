package main

import "encoding/json"

// StreamEvent is the envelope for iperf3 --json-stream output.
// Each line of stdout is one StreamEvent JSON object.
type StreamEvent struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

// StartData is the payload for the "start" event.
type StartData struct {
	Version   string `json:"version"`
	TestStart struct {
		Protocol  string `json:"protocol"`
		NumStreams int    `json:"num_streams"`
		Duration  int    `json:"duration"`
		Reverse   int    `json:"reverse"`
		Bidir     int    `json:"bidir"`
		Omit      int    `json:"omit"`
	} `json:"test_start"`
}

// IntervalSum is the per-interval aggregate (sum across all streams).
type IntervalSum struct {
	Start         float64 `json:"start"`
	End           float64 `json:"end"`
	BitsPerSecond float64 `json:"bits_per_second"`
	Bytes         int64   `json:"bytes"`
	Omitted       bool    `json:"omitted"`
	Sender        bool    `json:"sender"`
	// UDP fields
	Packets     int     `json:"packets"`
	LostPackets int     `json:"lost_packets"`
	LostPercent float64 `json:"lost_percent"`
	JitterMs    float64 `json:"jitter_ms"`
}

// IntervalData is the payload for the "interval" event.
type IntervalData struct {
	Streams []IntervalSum `json:"streams"`
	Sum     IntervalSum   `json:"sum"`
}

// EndSum is the final aggregate for sent or received.
type EndSum struct {
	Start         float64 `json:"start"`
	End           float64 `json:"end"`
	BitsPerSecond float64 `json:"bits_per_second"`
	Bytes         int64   `json:"bytes"`
	// UDP fields
	JitterMs    float64 `json:"jitter_ms"`
	LostPackets int     `json:"lost_packets"`
	LostPercent float64 `json:"lost_percent"`
	Packets     int     `json:"packets"`
}

// EndData is the payload for the "end" event.
type EndData struct {
	SumSent     EndSum `json:"sum_sent"`
	SumReceived EndSum `json:"sum_received"`
	CPUUtil     struct {
		HostTotal   float64 `json:"host_total"`
		RemoteTotal float64 `json:"remote_total"`
	} `json:"cpu_utilization_percent"`
}
