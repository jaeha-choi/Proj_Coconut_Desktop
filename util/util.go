package util

import (
	"errors"
	"github.com/jaeha-choi/Proj_Coconut_Utility/log"
	"io"
)

const (
	bufferSize = 4096
)

// readNString reads up to nth character. Maximum buffer size does not exceed bufferSize.
func readNString(reader io.Reader, n uint32) (string, error) {
	var buffSize uint32
	var totalReceived uint32 = 0
	resultString := ""

	if n < bufferSize {
		buffSize = n
	} else {
		buffSize = bufferSize
	}

	buffer := make([]byte, buffSize)

	var recv int
	var err error
	for totalReceived < n {
		if totalReceived+buffSize > n {
			buffer, err = io.ReadAll(io.LimitReader(reader, int64(n-totalReceived)))
			recv = len(buffer)
			if totalReceived+uint32(recv) != n {
				log.Warning("File not fully received")
				return "", errors.New("unexpected EOF")
			}
		} else {
			recv, err = io.ReadFull(reader, buffer)
		}
		if err != nil {
			log.Warning("Error while receiving string")
			return "", err
		}
		resultString += string(buffer)
		totalReceived += uint32(recv)
	}
	return resultString, nil
}

// readNString2 is similar to readNString except that there is no limit on buffer size.
func readNString2(reader io.Reader, n uint32) (string, error) {
	buffer, err := readNBytes(reader, n)
	return string(buffer), err
}

// readNBytes reads up to nth byte
func readNBytes(reader io.Reader, n uint32) ([]byte, error) {
	buffer := make([]byte, n)
	_, err := io.ReadFull(reader, buffer)
	return buffer, err
}
