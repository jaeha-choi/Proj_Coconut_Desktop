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
	//t.Cleanup(func() {
	//	pubFileN := "key.pub"
	//	privFileN := "key.priv"
	//	if err := os.Remove(pubFileN); err != nil {
	//		t.Error("Could not remove public key")
	//	}
	//	if err := os.Remove(privFileN); err != nil {
	//		t.Error("Could not remove private key")
	//	}
	//})
	bitSize := 1024
	key, err := createRSAKey(bitSize)
	if err != nil || key.N.BitLen() != bitSize {
		log.Debug(err)
		t.Error("Error in createRSAKey")
		return
	}
}

func TestOpenKeys(t *testing.T) {
	pubPem, privPem, err := OpenKeys()
	if pubPem == nil || privPem == nil || err != nil {
		log.Debug(err)
		t.Error("Error in OpenKeys")
		return
	}
}

func TestPemToKeys(t *testing.T) {
	pubPem, privPem, err := OpenKeys()
	if pubPem == nil || privPem == nil || err != nil {
		log.Debug(err)
		t.Error("Error in OpenKeys")
		return
	}

	if privKey, err := PemToKeys(privPem); privKey == nil || err != nil {
		log.Debug(err)
		log.Error("Error in PemToKeys")
		return
	}
}

func TestPemToSha256(t *testing.T) {
	pubPem, privPem, err := OpenKeys()
	if pubPem == nil || privPem == nil || err != nil {
		log.Debug(err)
		t.Error("Error in OpenKeys")
		return
	}
	PemToSha256(pubPem)
}
