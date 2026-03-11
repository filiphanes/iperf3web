//go:build !release

package main

// iperf3Executable returns the iperf3 binary to use.
// In non-release builds, rely on iperf3 being in PATH.
func iperf3Executable() string { return "iperf3" }
