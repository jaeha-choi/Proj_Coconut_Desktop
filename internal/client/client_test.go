package client

import (
	"github.com/jaeha-choi/Proj_Coconut_Utility/log"
	"os"
	"testing"
)

func init() {
	log.Init(os.Stdout, log.DEBUG)
}

func TestConnect(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Error(err)
	}
	err = client.Connect()
	if err != nil {
		t.Error(err)
		return
	}
}
