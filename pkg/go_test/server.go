package go_test

import (
	"fmt"
	"net"
	"time"
)

func WaitForTCPServer(host string, timeout time.Duration) error {
	timer := time.NewTimer(timeout)

	for {
		select {
		case <-timer.C:
			return fmt.Errorf("server [%s] did not start", host)
		default:
			_, err := net.DialTimeout("tcp", host, 50*time.Millisecond)
			if err == nil {
				return nil
			}
		}
	}
}
