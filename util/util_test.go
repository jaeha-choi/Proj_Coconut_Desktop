package util

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"fmt"
	"github.com/jaeha-choi/Proj_Coconut_Utility/log"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestInit(t *testing.T) {
	log.Init(os.Stdout, log.DEBUG)
}

func TestReadSizeError(t *testing.T) {
	inputBytes := make([]byte, 2)
	// [0 0]
	reader := bytes.NewReader(inputBytes)
	result, err := readSize(reader)
	if result != 0 || err == nil {
		t.Error("Expected error, but no error raised.")
	}
}

func TestReadSize0(t *testing.T) {
	inputBytes := make([]byte, 4)
	// [0 0 0 0]
	reader := bytes.NewReader(inputBytes)
	result, err := readSize(reader)
	if result != 0 || err != nil {
		t.Error("Error while reading size [0]")
	}
}

func TestReadSize255(t *testing.T) {
	inputBytes := make([]byte, 4)
	// [0 0 0 255]
	inputBytes[3] = 255
	reader := bytes.NewReader(inputBytes)
	result, err := readSize(reader)
	if result != 255 || err != nil {
		t.Error("Error while reading size [255]")
	}
}

func TestReadSize4095(t *testing.T) {
	inputBytes := make([]byte, 4)
	// [0 0 15 255]
	inputBytes[2] = 15
	inputBytes[3] = 255
	reader := bytes.NewReader(inputBytes)
	result, err := readSize(reader)
	if result != 4095 || err != nil {
		t.Error("Error while reading size [4095]")
	}
}

func TestReadSizeMax(t *testing.T) {
	inputBytes := make([]byte, 4)
	// [255 255 255 255]
	inputBytes[0] = 255
	inputBytes[1] = 255
	inputBytes[2] = 255
	inputBytes[3] = 255
	reader := bytes.NewReader(inputBytes)
	result, err := readSize(reader)
	if result != uint32Max || err != nil {
		t.Error("Error while reading size uint32max")
	}
}

func TestWriteSize0(t *testing.T) {
	result := writeSize(0)
	expected := make([]byte, 4)
	// [0 0 0 0]
	if bytes.Compare(result, expected) != 0 {
		log.Debug("Result: ", result)
		log.Debug("Expected: ", expected)
		t.Error("Incorrect value returned")
	}
}

func TestWriteSize255(t *testing.T) {
	result := writeSize(255)
	expected := make([]byte, 4)
	// [0 0 0 255]
	expected[3] = 255
	if bytes.Compare(result, expected) != 0 {
		log.Debug("Result: ", result)
		log.Debug("Expected: ", expected)
		t.Error("Incorrect value returned")
	}
}

func TestWriteSize4095(t *testing.T) {
	result := writeSize(4095)
	expected := make([]byte, 4)
	// [0 0 15 255]
	expected[2] = 15
	expected[3] = 255
	if bytes.Compare(result, expected) != 0 {
		log.Debug("Result: ", result)
		log.Debug("Expected: ", expected)
		t.Error("Incorrect value returned")
	}
}

func TestWriteSizeMax(t *testing.T) {
	result := writeSize(4294967295)
	expected := make([]byte, 4)
	// [255 255 255 255]
	expected[0] = 255
	expected[1] = 255
	expected[2] = 255
	expected[3] = 255
	if bytes.Compare(result, expected) != 0 {
		log.Debug("Result: ", result)
		log.Debug("Expected: ", expected)
		t.Error("Incorrect value returned")
	}
}

func TestReadNString(t *testing.T) {
	reader := strings.NewReader("init 1234 prev 5678 curr next \n")
	s, _ := readNString(reader, 22)
	if s != "init 1234 prev 5678 cu" {
		t.Error("Incorrect result.")
	}
}

func TestReadNStringBufferSize(t *testing.T) {
	// Open input file for testing
	file, err := os.Open("../testdata/log/long_text.txt")
	if err != nil {
		t.Error("Cannot read the input file")
	}

	// Close file when done
	defer func() {
		if file != nil {
			if err = file.Close(); err != nil {
				t.Error("Input file not properly closed")
			}
		}
	}()

	// Get sha-1 sum of original file
	h := sha1.New()
	if _, err := io.Copy(h, file); err != nil {
		t.Error("Error while getting shasum")
	}
	f1Hash := fmt.Sprintf("%x", h.Sum(nil))
	log.Info("Hash for input file: ", f1Hash)

	// Reset reader offset since the file was already read once
	_, err = file.Seek(0, 0)
	if err != nil {
		t.Error("Error while resetting reader offset")
	}

	// Test readNString
	reader := bufio.NewReader(file)
	s, err := readNString(reader, 4096)
	if err != nil {
		t.Error("readNString returned error")
	}

	// Create temp file
	tmpFile, err := ioutil.TempFile("", "tmp")
	if err != nil {
		t.Error("Error while creating temp file")
	}

	// Close and delete temp file when done
	defer func(name string) {
		if err := tmpFile.Close(); err != nil {
			t.Error("Temp file not closed.")
		}
		// Disable this statement to prevent the program from deleting
		// the file after testing
		if err := os.Remove(name); err != nil {
			log.Info("Temp file: ", name)
			t.Error("Temp file not removed.")
		}
	}(tmpFile.Name())

	// Write to tmp file for debugging
	if _, err := tmpFile.WriteString(s); err != nil {
		t.Error("Error while writing output file.")
	}

	// To guarantee the file to be written on disk before comparing
	if err = tmpFile.Sync(); err != nil {
		t.Error("File not written to disk")
	}

	// Reset reader offset since the file was already read once
	_, err = tmpFile.Seek(0, 0)
	if err != nil {
		t.Error("Error while setting reader offset")
	}

	// Get sha-1 sum of output file
	h2 := sha1.New()
	if _, err := io.Copy(h2, tmpFile); err != nil {
		t.Error("Error while getting shasum")
	}
	f2Hash := fmt.Sprintf("%x", h2.Sum(nil))
	log.Info("Hash for output file: ", f2Hash)

	if f1Hash != f2Hash {
		t.Error("checksum does not match")
	}
}

func TestReadNStringMaxSizePlugOne(t *testing.T) {
	// Open input file for testing
	file, err := os.Open("../testdata/log/long_text.txt")
	if err != nil {
		t.Error("Cannot read the input file")
	}

	// Close file when done
	defer func() {
		if file != nil {
			if err = file.Close(); err != nil {
				t.Error("Input file not properly closed")
			}
		}
	}()

	// Reset reader offset since the file was already read once
	_, err = file.Seek(0, 0)
	if err != nil {
		t.Error("Error while resetting reader offset")
	}

	// Test readNString
	reader := bufio.NewReader(file)
	s, err := readNString(reader, 4097)
	if s != "" || err == nil {
		t.Error("Expected error, but error was not raised")
	}
}

func TestReadNStringBufferSizeMinusOne(t *testing.T) {
	// Open input file for testing
	file, err := os.Open("../testdata/log/long_text.txt")
	if err != nil {
		t.Error("Cannot read the input file")
	}

	// Close file when done
	defer func() {
		if file != nil {
			if err = file.Close(); err != nil {
				t.Error("Input file not properly closed")
			}
		}
	}()

	inputReader := io.LimitReader(file, 4095)

	// Get sha-1 sum of original file
	h := sha1.New()
	if _, err := io.Copy(h, inputReader); err != nil {
		t.Error("Error while getting shasum")
	}
	f1Hash := fmt.Sprintf("%x", h.Sum(nil))
	log.Info("Hash for input file: ", f1Hash)

	// Reset reader offset since the file was already read once
	_, err = file.Seek(0, 0)
	if err != nil {
		t.Error("Error while resetting reader offset")
	}

	// Test readNString
	reader := bufio.NewReader(file)
	s, _ := readNString(reader, 4095)

	tmpReader := strings.NewReader(s[:4095])

	// Get sha-1 sum of output file
	h2 := sha1.New()
	if _, err := io.Copy(h2, tmpReader); err != nil {
		t.Error("Error while getting shasum")
	}
	f2Hash := fmt.Sprintf("%x", h2.Sum(nil))
	log.Info("Hash for output: ", f2Hash)

	if f1Hash != f2Hash {
		t.Error("checksum does not match")
	}
}

func TestReadNStringZero(t *testing.T) {
	reader := strings.NewReader("init 1234 prev 5678 curr next \n")
	s, _ := readNString(reader, 0)
	if s != "" {
		t.Error("Incorrect result.")
	}
}

func TestReadNStringExceed(t *testing.T) {
	reader := strings.NewReader("init 1234 prev 5678 curr next \n")
	_, err := readNString(reader, 50)
	if err == nil {
		t.Error("Expected error, but no error raised.")
	}
}

func TestIntToUint32SignedInt(t *testing.T) {
	val, err := IntToUint32(-1)
	if val != 0 || err == nil {
		t.Error("Expected error, but no error raised.")
	}
}

func TestIntToUint32Max(t *testing.T) {
	val, err := IntToUint32(4294967296)
	if val != 0 || err == nil {
		t.Error("Expected error, but no error raised.")
	}
}

func TestIntToUint32(t *testing.T) {
	val, err := IntToUint32(27532)
	if val != 27532 || err != nil {
		t.Error("Error during conversion")
	}
}
