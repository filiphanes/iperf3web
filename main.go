package main

import (
	"embed"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"
)

//go:embed static
var staticFiles embed.FS

func findFreePort() int {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 8765
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func openBrowser(url string) {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd, args = "open", []string{url}
	case "windows":
		cmd, args = "cmd", []string{"/c", "start", url}
	default:
		cmd, args = "xdg-open", []string{url}
	}
	if err := exec.Command(cmd, args...).Start(); err != nil {
		log.Printf("could not open browser: %v", err)
	}
}

func main() {
	iperf3Path, err := exec.LookPath("iperf3")
	if err != nil {
		log.Println("WARNING: iperf3 not found in PATH")
		log.Println("  macOS:   brew install iperf3")
		log.Println("  Ubuntu:  sudo apt install iperf3")
		log.Println("  Windows: https://iperf.fr/iperf-download.php")
		log.Println("Tests will fail until iperf3 is installed and in PATH.")
	} else {
		log.Printf("Found iperf3 at: %s", iperf3Path)
	}

	port := findFreePort()
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	url := fmt.Sprintf("http://%s", addr)

	history := newHistoryStore("iperf3web-history.json")
	runner := newRunner()
	setupRoutes(runner, history)

	log.Printf("iperf3 web GUI running at %s", url)
	go func() {
		time.Sleep(300 * time.Millisecond)
		openBrowser(url)
	}()

	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
