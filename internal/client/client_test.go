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
	Connect()
}
