package cryptography

import (
	"github.com/jaeha-choi/Proj_Coconut_Utility/log"
	"io/ioutil"
	"os"
	"testing"
)

func TestReadTest(t *testing.T) {
	testFileN := "../testdata/checksum.txt"
	_, privPem, err := OpenKeys("./")
	if err != nil {
		log.Debug(err)
		t.Error("Error in OpenKeys")
		return
	}
	privKey, err := PemToKeys(privPem)
	if err != nil {
		log.Debug(err)
		t.Error("Error in PemToKeys")
		return
	}

	tmpFile, err := ioutil.TempFile(".", "test")
	if err != nil {
		log.Debug(err)
		t.Error("Error opening temp file")
		return
	}
	defer func() {
		if err := tmpFile.Close(); err != nil {
			log.Debug(err)
			t.Error("Error closing tmp file")
			return
		}
		if err := os.Remove(tmpFile.Name()); err != nil {
			log.Debug(err)
			t.Error("Error while removing temp file. Temp file at: ", tmpFile.Name())
			return
		}
	}()

	// Encrypt
	streamEncrypt, err := EncryptSetup(testFileN)
	if err != nil {
		log.Debug(err)
		t.Error("Error in EncryptSetup")
		return
	}
	err = streamEncrypt.Encrypt(tmpFile, &privKey.PublicKey, privKey)
	if err != nil {
		log.Debug(err)
		t.Error("Error in Encrypt")
		return
	}
	defer func() {
		if err := streamEncrypt.Close(); err != nil {
			log.Debug(err)
			t.Error("Error in Close")
			return
		}
	}()

	// Reset offset
	if _, err := tmpFile.Seek(0, 0); err != nil {
		log.Debug(err)
		t.Error("Error resetting offset")
		return
	}

	// Decrypt
	streamDecrypt, err := DecryptSetup()
	if err != nil {
		log.Debug(err)
		t.Error("Error in DecryptSetup")
		return
	}
	err = streamDecrypt.Decrypt(tmpFile, &privKey.PublicKey, privKey)
	if err != nil {
		log.Debug(err)
		t.Error("Error in Decrypt")
		return
	}
	defer func() {
		if err := streamDecrypt.Close(); err != nil {
			log.Debug(err)
			t.Error("Error in Close")
			return
		}
	}()
}
