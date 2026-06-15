package health

import (
	"net"
	"strconv"
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
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return StatusUnreachable
	}
	conn.Close()
	return StatusReachable
}
