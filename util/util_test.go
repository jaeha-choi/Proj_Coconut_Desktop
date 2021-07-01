package util

import "testing"

func TestCheckIPAddress(t *testing.T) {
	validIPV4 := "10.40.210.253"
	CheckIPAddress(validIPV4)
	if CheckIPAddress(validIPV4) == false {
		t.Error("IP Address is invalid")
	}

	invalidIPV4 := "1000.40.210.253"
	CheckIPAddress(invalidIPV4)
	if CheckIPAddress(invalidIPV4) == false {
		t.Error("IP Address is invalid")
	}
	validIPV6 := "2001:0db8:85a3:0000:0000:8a2e:0370:7334"
	CheckIPAddress(validIPV6)
	if CheckIPAddress(validIPV6) == false {
		t.Error("IP Address is invalid")
	}

	invalidIPV6 := "2001:0db8:85a3:0000:0000:8a2e:0370:7334:3445"
	if CheckIPAddress(invalidIPV6) == false {
		t.Error("IP Address is invalid")
	}
}
