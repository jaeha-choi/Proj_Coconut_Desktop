package util

import (
	"net"
)

// TestCheckIPAddress check if ip address is valid or not
func CheckIPAddress(ip string) bool {
	if net.ParseIP(ip) == nil {
		//log.Error("IP Address: " + ip + "- Invalid\n")
		return false
	} else {
		//log.Error("IP Address: " + ip + " - Valid\n")
		return true
	}
}
