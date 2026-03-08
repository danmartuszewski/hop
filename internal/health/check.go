package health

import (
	"fmt"
	"net"
	"time"
)

type Status int

const (
	StatusUnknown     Status = iota
	StatusChecking
	StatusReachable
	StatusUnreachable
)

func CheckTCP(host string, port int) Status {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return StatusUnreachable
	}
	conn.Close()
	return StatusReachable
}
