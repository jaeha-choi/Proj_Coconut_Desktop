package util

import (
	"encoding/binary"
	"errors"
	"github.com/jaeha-choi/Proj_Coconut_Utility/log"
	"io"
	"net"
)

const (
	bufferSize = 4096
)

// ReadString reads string from a connection
func ReadString(conn net.Conn) (string, error) {
	// Read packet size (string size)
	size, err := readSize(conn)
	if err != nil {
		log.Error("Error while reading string size")
		return "", err
	}
	// Read string from the packet
	str, err := readNString(conn, size)
	if err != nil {
		log.Error("Error while reading string")
		return "", err
	}
	return str, nil
}

// readSize reads first 4 bytes from the reader and convert them into a uint32 value
func readSize(reader io.Reader) (uint32, error) {
	// Read first 4 bytes for the size
	b, err := readNBytes(reader, 4)
	if err != nil {
		log.Error("Error while reading packet size")
		return 0, err
	}
	return binary.BigEndian.Uint32(b), nil
}

// writeSize converts unsigned integer 32 to bytes
// Take a look at encoding/gob package or protocol buffers for a better performance.
func writeSize(size uint32) []byte {
	// consider using array over slice for a better performance i.e: arr := [4]byte{}
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, size)
	return b
}

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
