package health

import (
	"net"
	"testing"
)

func TestCheckTCP_Reachable(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	defer ln.Close()

	port := ln.Addr().(*net.TCPAddr).Port

	status := CheckTCP("127.0.0.1", port)
	if status != StatusReachable {
		t.Errorf("expected StatusReachable, got %d", status)
	}
}

func TestCheckTCP_Unreachable(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	status := CheckTCP("127.0.0.1", port)
	if status != StatusUnreachable {
		t.Errorf("expected StatusUnreachable, got %d", status)
	}
}
