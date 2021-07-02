package util

import (
	"net"
)

// TestCheckIPAddress check if ip address is valid or not
func CheckIPAddress(ip string) bool {
	return net.ParseIP(ip) != nil
}
