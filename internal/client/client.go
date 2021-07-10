package client

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"github.com/jaeha-choi/Proj_Coconut_Utility/log"
	"io/ioutil"
	"os"
)

func createRSAKey(bitSize int) (*rsa.PrivateKey, error) {
	reader := rand.Reader
	key, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		log.Debug(err)
		log.Error("Error while creating RSA key")
		return nil, err
	}
	return key, nil
}

func OpenKeys() (*pem.Block, *pem.Block, error) {
	var err error
	var pubBlock *pem.Block
	var privBlock *pem.Block
	var newKeyBitSize = 4096
	pubFileN := "key.pub"
	privFileN := "key.priv"

	// Check if keys already exist
	if _, err := os.Stat(pubFileN); !os.IsNotExist(err) {

		// Read existing pubkey file
		pubFileBytes, err := ioutil.ReadFile(pubFileN)
		if err != nil {
			log.Debug(err)
			log.Error("Error while opening RSA pubkey in OpenKeys")
			return nil, nil, err
		}
		// Decode file and create block
		pubBlock, _ = pem.Decode(pubFileBytes)
		if pubBlock == nil {
			log.Error("Error while decoding pubkey file")
			return nil, nil, err
		}

		// Read existing privkey file
		privFileBytes, err := ioutil.ReadFile(privFileN)
		if err != nil {
			log.Debug(err)
			log.Error("Error while opening RSA privfile in OpenKeys")
			return nil, nil, err
		}
		// Decode file and create block
		privBlock, _ = pem.Decode(privFileBytes)
		if privBlock == nil {
			log.Error("Error while decoding priv file")
			return nil, nil, err
		}
	} else if os.IsNotExist(err) {
		log.Debug(err)
		// Create new RSA key pair
		key, err := createRSAKey(newKeyBitSize)
		if err != nil {
			log.Debug(err)
			log.Error("Error while creating RSA key in OpenKeys")
			return nil, nil, err
		}
		// Open pubkey file
		pubOut, err := os.Create(pubFileN)
		if err != nil {
			log.Debug(err)
			log.Error("Error while creating pubkey file")
			return nil, nil, err
		}
		// Close pubkey file when done
		defer func() {
			if errDef := pubOut.Close(); errDef != nil {
				log.Error(err)
				log.Error("Error while closing pubkey file")
				err = errDef
				return
			}
			if errDef := os.Chmod(pubFileN, 0644); errDef != nil {
				log.Debug(errDef)
				log.Error("Error while updating pubkey file permissions")
				err = errDef
				return
			}
		}()
		// Create PEM block
		pubBlock = &pem.Block{
			Type:    "RSA PUBLIC KEY",
			Headers: nil,
			Bytes:   x509.MarshalPKCS1PublicKey(&key.PublicKey),
		}
		// Write PEM block to pubkey file
		if err := pem.Encode(pubOut, pubBlock); err != nil {
			log.Debug(err)
			log.Error("Error while writing to pubkey file")
			return nil, nil, err
		}

		// Create privkey file
		privOut, err := os.OpenFile(privFileN, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
		if err != nil {
			log.Debug(err)
			log.Error("Error while creating private key in OpenKeys")
			return nil, nil, err
		}
		// Close privkey file when done
		defer func() {
			if errDef := privOut.Close(); errDef != nil {
				log.Debug(errDef)
				log.Error("Error while closing private key file")
				err = errDef
				return
			}
			if errDef := os.Chmod(privFileN, 0600); errDef != nil {
				log.Debug(errDef)
				log.Error("Error while updating permissions")
				err = errDef
				return
			}
		}()
		// Create PEM block
		privBlock = &pem.Block{
			Type:    "RSA PRIVATE KEY",
			Headers: nil,
			Bytes:   x509.MarshalPKCS1PrivateKey(key),
		}
		// Write PEM block to privkey file
		if err := pem.Encode(privOut, privBlock); err != nil {
			log.Debug(err)
			log.Error("Error while encoding private key")
			return nil, nil, err
		}
	} else {
		log.Debug(err)
		log.Error("Error while checking stat of the file")
	}
	return pubBlock, privBlock, err
}

// PemToKeys convert private PEM block to rsa.PrivateKey struct.
func PemToKeys(privBlock *pem.Block) (*rsa.PrivateKey, error) {
	key, err := x509.ParsePKCS1PrivateKey(privBlock.Bytes)
	if err != nil {
		log.Debug(err)
		log.Error("Error while converting PEM block to keys")
		return nil, err
	}
	return key, nil
}

// PemToSha256 generates bytes containing sha256sum of pubBlock bytes.
func PemToSha256(pubBlock *pem.Block) []byte {
	// sha256sum always returns 32 bytes
	hash := sha256.Sum256(pubBlock.Bytes)

	// Base64 seems unnecessary as of now, but in case raw bytes cause issues,
	// enable them by uncommenting following codes.

	// Base64 of sha256sum would always generate 44 bytes including a padding,
	// but in case we change hash method from sha256, I won't hardcode it.
	//encoded := make([]byte, base64.StdEncoding.EncodedLen(len(hash)))
	//base64.StdEncoding.Encode(encoded, hash[:])
	//return encoded
	return hash[:]
}
