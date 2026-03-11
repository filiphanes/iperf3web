//go:build release

package main

import _ "embed"

//go:embed embedded/iperf3-darwin-arm64
var embeddedIperf3 []byte

func iperf3Executable() string { return extractIperf3(embeddedIperf3, "") }
