package util

import (
	"github.com/jaeha-choi/Proj_Coconut_Utility/log"
	"os"
	"strings"
	"testing"
)

func TestInit(t *testing.T) {
	log.Init(os.Stdout, log.DEBUG)
}

func TestReadNString(t *testing.T) {
	reader := strings.NewReader("init 1234 prev 5678 curr next \n")
	s, _ := readNString(reader, 22)
	if s != "init 1234 prev 5678 cu" {
		t.Error("Incorrect result.")
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
		t.Error("Excepted error, but no error raised.")
	}
}

func TestReadNString2(t *testing.T) {
	reader := strings.NewReader("init 1234 prev 5678 curr next \n")
	s, _ := readNString2(reader, 22)
	if s != "init 1234 prev 5678 cu" {
		t.Error("Incorrect result.")
	}
}

func TestReadNString2Exceed(t *testing.T) {
	reader := strings.NewReader("init 1234 prev 5678 curr next \n")
	_, err := readNString2(reader, 50)
	if err == nil {
		t.Error("Excepted error, but no error raised.")
	}
}

func TestReadNString2Zero(t *testing.T) {
	reader := strings.NewReader("init 1234 prev 5678 curr next \n")
	s, _ := readNString(reader, 0)
	if s != "" {
		t.Error("Incorrect result.")
	}
}
