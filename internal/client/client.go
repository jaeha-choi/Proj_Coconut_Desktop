package client

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"github.com/jaeha-choi/Proj_Coconut_Utility/log"
	"os"
)

var ExistingKeyError = errors.New("keys already exist")

func CreateRSAKey(bitSize int) (*rsa.PrivateKey, error) {
	reader := rand.Reader
	key, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		log.Debug(err)
		log.Error("Error while creating RSA key")
		return nil, err
	}
	return key, nil
}

func SaveRSAKeys(key *rsa.PrivateKey) error {
	pubFileN := "key.pub"
	privFileN := "key.priv"
	var err error = nil

	// Check if keys already exist
	if _, err := os.Stat(pubFileN); err == nil {
		log.Debug(err)
		return ExistingKeyError
	}
	// Open pubkey file
	pubOut, err := os.Create(pubFileN)
	if err != nil {
		log.Debug(err)
		log.Error("Error while opening file")
		return err
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
			log.Error("Error while updating permissions")
			err = errDef
			return
		}
	}()
	// Create PEM block
	block := pem.Block{
		Type:    "RSA PUBLIC KEY",
		Headers: nil,
		Bytes:   x509.MarshalPKCS1PublicKey(&key.PublicKey),
	}
	// Write to pubkey file
	err = pem.Encode(pubOut, &block)
	if err != nil {
		log.Debug(err)
		log.Error("Error while encoding public key")
		return err
	}

	// Create priv key file
	privOut, err := os.OpenFile(privFileN, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		log.Debug(err)
		log.Error("Error while creating private key")
		return err
	}
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
	block = pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   x509.MarshalPKCS1PrivateKey(key),
	}
	// Write to priv file
	err = pem.Encode(privOut, &block)
	if err != nil {
		log.Debug(err)
		log.Error("Error while encoding private key")
		return err
	}

	return err
}
