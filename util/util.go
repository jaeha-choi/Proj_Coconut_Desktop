package util

import (
	"net"
)

// CheckIPAddress check if ip address is valid or not
func CheckIPAddress(ip string) bool {
	return net.ParseIP(ip) != nil
}
