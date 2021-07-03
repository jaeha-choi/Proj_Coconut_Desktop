package util

import (
	"encoding/binary"
	"errors"
	"github.com/jaeha-choi/Proj_Coconut_Utility/log"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	bufferSize   = 4096
	uint32Max    = 4294967295
	downloadPath = "./downloaded"
)

// ReadString reads string from a connection
func ReadString(reader io.Reader) (string, error) {
	// Read packet size (string size)
	size, err := readSize(reader)
	if err != nil {
		log.Debug(err)
		log.Error("Error while reading string size")
		return "", err
	}

	// ReadString always expect the size to be <= bufferSize
	if size > bufferSize {
		log.Error("String size cannot be greater than ", bufferSize, ". String size: ", size)
		return "", errors.New("size exceeded")
	}

	// Read string from the packet
	str, err := readNString(reader, size)
	if err != nil {
		log.Debug(err)
		log.Error("Error while reading string")
		return "", err
	}
	return str, nil
}

// ReadBinary reads file name and file content from a connection and save it.
func ReadBinary(reader io.Reader) error {
	// Read file name
	fileN, err := ReadString(reader)
	if fileN == "" || err != nil {
		log.Debug(err)
		log.Error("Error while reading file name")
		return err
	}

	// Read file size
	size, err := readSize(reader)
	if err != nil {
		log.Debug(err)
		log.Error("Error while file size")
		return err
	}

	// Read file and save
	if err := readNBinary(reader, size, fileN); err != nil {
		log.Debug(err)
		log.Error("Error while reading/saving binary file")
		return err
	}
	return nil
}

// readSize reads first 4 bytes from the reader and convert them into a uint32 value
func readSize(reader io.Reader) (uint32, error) {
	// Read first 4 bytes for the size
	b, err := readNBytes(reader, 4)
	if err != nil {
		log.Debug(err)
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

// readNBinary reads up to nth bytes and save it as fileN in downloadPath.
// Maximum buffer size does not exceed bufferSize.
// Returns error == nil only if file is fully downloaded, renamed and moved to downloadPath.
func readNBinary(reader io.Reader, n uint32, fileN string) error {
	var buffSize uint32
	var totalReceived uint32 = 0
	var isDownloadComplete = false
	var receivedLen int

	if n < bufferSize {
		buffSize = n
	} else {
		buffSize = bufferSize
	}

	buffer := make([]byte, buffSize)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(downloadPath, os.ModePerm); err != nil {
		log.Debug(err)
		log.Error("Error while creating download directory")
		return err
	}

	// Create temporary file for downloading
	tmpFile, err := ioutil.TempFile(downloadPath, ".tmp_download_")
	if err != nil {
		log.Debug(err)
		log.Error("Temp file could not be opened")
	}

	// Close and delete temp file when done
	defer func(name string) {
		if !isDownloadComplete {
			if err := tmpFile.Close(); err != nil {
				log.Debug(err)
				log.Error("Error while closing the file.")
			}
			// Delete file if not renamed
			if err := os.Remove(name); err != nil {
				log.Debug(err)
				log.Error("Error while removing temp file. Temp file at: ", name)
			}
		}
	}(tmpFile.Name())

	// Repeat downloading until the file is fully received
	for totalReceived < n {
		if totalReceived+buffSize > n {
			buffer, err = io.ReadAll(io.LimitReader(reader, int64(n-totalReceived)))
			receivedLen = len(buffer)
			// If reader contains less than expected n
			if totalReceived+uint32(receivedLen) != n {
				log.Error("File not fully received")
				return errors.New("unexpected EOF")
			}
		} else {
			receivedLen, err = io.ReadFull(reader, buffer)
		}

		if err != nil {
			log.Debug(err)
			log.Error("Error while receiving bytes")
			return err
		}

		writtenLen, err := tmpFile.Write(buffer)

		// If error encountered while writing a file, close then delete tmp file.
		if writtenLen != receivedLen || err != nil {
			log.Debug(err)
			log.Error("Error while writing to a file")
			return err
		}
		totalReceived += uint32(receivedLen)
	}
	isDownloadComplete = true

	// Close I/O operation for temporary file
	if err := tmpFile.Close(); err != nil {
		log.Debug(err)
		log.Error("Error while closing temp file.")
		return err
	}

	// Move temporary file to download directory (downloadPath)
	if err := os.Rename(tmpFile.Name(), filepath.Join(downloadPath, fileN)); err != nil {
		log.Debug(err)
		log.Error("Error moving the temp file to download path")
		if err := os.Remove(tmpFile.Name()); err != nil {
			log.Debug(err)
			log.Error("Error while removing temp file. Temp file at: ", tmpFile.Name())
		}
		return err
	}
	return nil
}

// readNString reads up to nth character. Maximum length should not exceed bufferSize.
func readNString(reader io.Reader, n uint32) (string, error) {
	if n > bufferSize {
		log.Error("n should be smaller than ", bufferSize)
		return "", errors.New("parameter value error")
	}
	buffer, err := readNBytes(reader, n)
	return string(buffer), err
}

// readNBytes reads up to nth byte
func readNBytes(reader io.Reader, n uint32) ([]byte, error) {
	buffer := make([]byte, n)
	_, err := io.ReadFull(reader, buffer)
	return buffer, err
}

// IntToUint32 converts integer value to uint32. Returns error if value occurs overflow
func IntToUint32(n int) (uint32, error) {
	if n < 0 || n > uint32Max {
		log.Error("value ", n, " overflows uint32")
		return 0, errors.New("value overflows uint32")
	}
	return uint32(n), nil
}
