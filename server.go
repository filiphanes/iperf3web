package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
)

func setupRoutes(runner *Runner, history *HistoryStore) {
	static, _ := fs.Sub(staticFiles, "static")
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(static))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		data, _ := staticFiles.ReadFile("static/index.html")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})

	http.HandleFunc("/api/servers", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, publicServers)
	})

	http.HandleFunc("/api/test/status", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]interface{}{
			"running": runner.isRunning(),
			"params":  runner.getParams(),
		})
	})

	http.HandleFunc("/api/test/start", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var params TestParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(params.Server) == "" {
			http.Error(w, "server is required", http.StatusBadRequest)
			return
		}
		// Apply defaults
		if params.Duration == 0 {
			params.Duration = 10
		}
		if params.Port == 0 {
			params.Port = 5201
		}
		if params.Protocol == "" {
			params.Protocol = "tcp"
		}
		if params.Direction == "" {
			params.Direction = "upload"
		}
		if params.Parallel == 0 {
			params.Parallel = 1
		}
		if err := runner.start(params, func(done testDone) {
			if done.errStr == "" { // don't save failed/cancelled tests
				history.add(newEntry(done))
			}
		}); err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		writeJSON(w, map[string]string{"status": "started"})
	})

	http.HandleFunc("/api/test/stop", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		runner.stop()
		writeJSON(w, map[string]string{"status": "stopped"})
	})

	// SSE stream — stays open for the life of the page.
	// All test events are pushed here.
	http.HandleFunc("/api/test/stream", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming not supported", http.StatusInternalServerError)
			return
		}

		ch := runner.subscribe()
		defer runner.unsubscribe(ch)

		// Send current status immediately so the client can sync on reconnect.
		sendSSE(w, flusher, SSEMsg{Type: "status", Payload: map[string]interface{}{
			"running": runner.isRunning(),
			"params":  runner.getParams(),
		}})

		for {
			select {
			case msg := <-ch:
				sendSSE(w, flusher, msg)
			case <-r.Context().Done():
				return
			}
		}
	})

	http.HandleFunc("/api/history", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			entries := history.getAll()
			if entries == nil {
				entries = []HistoryEntry{}
			}
			writeJSON(w, entries)
		case http.MethodDelete:
			history.clear()
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/history/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id := strings.TrimPrefix(r.URL.Path, "/api/history/")
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
		history.delete(id)
		w.WriteHeader(http.StatusNoContent)
	})
}

func sendSSE(w http.ResponseWriter, flusher http.Flusher, msg SSEMsg) {
	data, _ := json.Marshal(msg)
	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
