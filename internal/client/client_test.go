package client

import (
	"github.com/jaeha-choi/Proj_Coconut_Utility/cryptography"
	"github.com/jaeha-choi/Proj_Coconut_Utility/util"
	"testing"
	//"time"
)

func initClient() Client {
	client := InitConfig()
	client.ServerHost = "coconut-demo.jaeha.dev"
	pubBlock, _ := cryptography.OpenKeysAsBlock(client.KeyPath, "key.pub")
	privBlock, _ := cryptography.OpenPrivKey(client.KeyPath, "key.priv")
	client.pubKeyBlock = pubBlock
	client.privKey = privBlock
	return *client
}

func TestDoOpenHolePunch(t *testing.T) {
	client := initClient()
	defer func() {
		_ = client.Disconnect()
	}()
	err := client.Connect()
	var key string
	key = "5P3jG+vfufBnwSxEuhrZei6H4ZbHuKCd/XnqpvLLMSI="
	err = client.DoRequestP2P([]byte(key))
	if err != nil {
		t.Error(err)
	}

}

func TestDoOpenHolePunch2(t *testing.T) {
	client := initClient()
	defer func() {
		_ = client.Disconnect()
	}()
	err := client.Connect()
	_, _ = util.ReadMessage(client.conn)
	err = client.handleRequestP2P()
	if err != nil {
		t.Error(err)
	}

}

//func TestConnect(t *testing.T) {
//	client, err := InitConfig()
//	if err != nil {
//		t.Error(err)
//	}
//	if err != nil {
//		return
//	}
//	log.Debug("init")
//	pubBlock, privBlock, err := cryptography.OpenKeys(client.KeyPath)
//	if err != nil {
//		log.Fatal(err)
//		os.Exit(1)
//	}
//	client.pubKeyBlock = pubBlock
//	client.privKey, err = cryptography.PemToKeys(privBlock)
//	if err != nil {
//		log.Fatal(err)
//		os.Exit(1)
//	}
//	defer func() {
//		err = client.Disconnect()
//	}()
//
//	err = client.Connect()
//	if err != nil {
//		t.Error(err)
//		return
//	}
//
//	log.Debug("connect")
//
//	if err != nil {
//		t.Error(err)
//		return
//	}
//	err = client.DoRequestP2P(client.conn, []byte(""))
//	if err != nil {
//		log.Error(err)
//	}
//	log.Debug("holepunch")
//}

//func TestGOBReadWrite(t *testing.T) {
//	client, err := InitConfig()
//	if err != nil {
//		t.Error(err)
//	}
//	client.DataPath = "./data/"
//	err = os.Remove("./data/contacts.gob")
//	if err != nil {
//		return
//	}
//
//	log.Debug("config")
//	err = client.ReadContactsFile()
//	if err != nil {
//		t.Error(err)
//	}
//	if client.contactMap != nil {
//		t.Error("Error initializing contact list")
//	}
//	log.Debug("file read")
//	var bytes1 = []byte("string")
//	var bytes2 = []byte("123456")
//	var bytes3 = []byte("abcdef")
//	var bytes4 = []byte("zyxwvu")
//	var pub = client.pubKeyBlock
//	if !client.addContact("abc", "zyx", bytes1, pub) {
//		t.Error("failed to add contact")
//	}
//	if !client.addContact("bbc", "dsx", bytes2, pub) {
//		t.Error("failed to add contact")
//	}
//	if !client.addContact("cbc", "fds", bytes3, pub) {
//		t.Error("failed to add contact")
//	}
//	if !client.addContact("dbc", "457", bytes4, pub) {
//		t.Error("failed to add contact")
//	}
//	log.Debug("added contacts")
//
//	if err = client.WriteContactsFile(); err != nil {
//		t.Error(err)
//	}
//	log.Debug("file write")
//	client.contactMap = nil
//	err = client.ReadContactsFile()
//	if client.contactMap == nil {
//		t.Error("Error opening contacts file")
//	}
//	log.Debug("file read 2")
//	var b bool
//	var c Contact
//	b, c = client.findContact([]byte("123456"))
//	if b {
//		log.Debug(c)
//	}
//	if !b {
//		log.Debug("contact not found")
//	}
//
//	b, c = client.findContact([]byte("123556"))
//	if b {
//		log.Debug(c)
//	}
//	if !b {
//		log.Debug("contact not found")
//	}
//
//	b, c = client.findContact([]byte("abcdef"))
//	if b {
//		log.Debug(c)
//	}
//	if !b {
//		log.Debug("contact not found")
//	}
//
//	if client.RemoveContact([]byte("123456")) {
//		log.Debug("successfully removed")
//	} else {
//		log.Debug("could not find contact")
//	}
//	if client.RemoveContact([]byte("123556")) {
//		log.Debug("successfully removed")
//	} else {
//		log.Debug("could not find contact")
//	}
//	if client.RemoveContact([]byte("abcdef")) {
//		log.Debug("successfully removed")
//	} else {
//		log.Debug("could not find contact")
//	}
//	log.Debug(client.contactMap)
//}
