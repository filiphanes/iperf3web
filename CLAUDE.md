# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```sh
go build -o iperf3web .   # build binary (~8MB, stdlib only, no deps)
./iperf3web               # starts server on random port, opens browser
```

Requires iperf3 3.6+ in PATH (uses `--json-stream` and `--forceflush` flags).
- macOS: `brew install iperf3` (installs to `/opt/homebrew/bin/iperf3`)

No tests exist yet. To verify the build works, run `go vet ./...`.

**Avoid running the binary from a zsh shell snapshot** — it floods stderr. Use `python3 -c "import subprocess; subprocess.run(['./iperf3web'])"` or a separate terminal.

## Architecture

Single-process Go server with an embedded static frontend. No external dependencies — stdlib only.

**Data flow:**
1. Browser POSTs `TestParams` to `/api/test/start`
2. `Runner` spawns `iperf3` subprocess with `--json-stream --forceflush`
3. `runner.go` reads newline-delimited JSON events from stdout (`start`, `interval`, `end`, `error`)
4. Each event is broadcast via `Runner.broadcast()` to all SSE subscribers
5. Browser receives events on `/api/test/stream` (persistent SSE connection, one per tab)
6. On completion, `HistoryStore` persists the result to `iperf3web-history.json` in CWD

**Key types:**
- `Runner` (`runner.go`) — mutex-protected subprocess manager + SSE fan-out hub. Only one test runs at a time. Clients subscribe/unsubscribe via channels; slow clients are dropped.
- `HistoryStore` (`history.go`) — RWMutex-protected in-memory list backed by JSON file. Capped at 500 entries. IDs are Unix millisecond timestamps.
- `TestParams` (`runner.go`) — all iperf3 options; `buildArgs()` translates to CLI flags.
- `StreamEvent` / `StartData` / `IntervalData` / `EndData` (`parser.go`) — iperf3 JSON stream structs.
- `SSEMsg` — `{type, payload}` JSON sent over the SSE stream. Types: `status`, `start`, `interval`, `end`, `error`, `done`.

**Frontend** (`static/index.html`): single HTML file, Bootstrap 5 + Chart.js from CDN, all JS inline. Connects to SSE on load and reconciles state on reconnect via the `status` event sent immediately upon subscription.

**Concurrency notes:**
- `Runner` uses two separate mutexes: `mu` for running/params state, `clientMu` for the SSE subscriber map.
- `HistoryStore` uses `sync.RWMutex`; `save()` is called after every mutation.
