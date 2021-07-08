package client

import (
	"github.com/jaeha-choi/Proj_Coconut_Utility/log"
	"os"
	"testing"
)

func init() {
	log.Init(os.Stdout, log.DEBUG)
}

func TestCreateRSAKey(t *testing.T) {
	t.Cleanup(func() {
		pubFileN := "key.pub"
		privFileN := "key.priv"
		if err := os.Remove(pubFileN); err != nil {
			t.Error("Could not remove public key")
		}
		if err := os.Remove(privFileN); err != nil {
			t.Error("Could not remove private key")
		}
	})
	bitSize := 1024
	key, err := CreateRSAKey(bitSize)
	if err != nil || key.N.BitLen() != bitSize {
		log.Debug(err)
		t.Error("Error in CreateRSAKey")
		return
	}
	if err := SaveRSAKeys(key); err != nil {
		log.Debug(err)
		t.Error("Error in SaveRSAKeys")
		return
	}
}
