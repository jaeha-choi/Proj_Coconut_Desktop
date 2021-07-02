package util

import "testing"

func TestCheckIPAddressValidIpv4(t *testing.T) {
	validIPV4 := "10.40.210.253"
	if !CheckIPAddress(validIPV4) {
		t.Error("TestCheckIPAddressValidIpv4 is invalid")
	}
}

func TestCheckIPAddressInvalidIpv4(t *testing.T) {
	invalidIPV4 := "1000.40.210.253"
	if CheckIPAddress(invalidIPV4) {
		t.Error("TestCheckIPAddressInvalidIpv4 is invalid")
	}
}

func TestCheckIPAddressValidIpv6(t *testing.T) {
	validIPV6 := "2001:0db8:85a3:0000:0000:8a2e:0370:7334"
	if !CheckIPAddress(validIPV6) {
		t.Error("TestCheckIPAddressValidIpv6 is invalid")
	}
}

func TestCheckIPAddressInvalidIpv6(t *testing.T) {
	invalidIPV6 := "2001:0db8:85a3:0000:0000:8a2e:0370:7334:3445"
	if CheckIPAddress(invalidIPV6) {
		t.Error("TestCheckIPAddressInvalidIpv6 is invalid")
	}
}
