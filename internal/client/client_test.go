package client

import (
	"github.com/jaeha-choi/Proj_Coconut_Utility/cryptography"
	"github.com/jaeha-choi/Proj_Coconut_Utility/log"
	"os"
	"testing"
)

func init() {
	log.Init(os.Stdout, log.DEBUG)
}

//TODO: Test with demo server
// both clients connect to server
// client1 adds client 2 via add code
// client1 sends P2P request to server
// server asks for pubkey hash
// client1 hashes client2 pubkey and sends to server
// client2 receives client1 local and remote ip/port
// client1 receives client2 local and remote ip/port
// client1 and client2 send init packets to both locations
// client1 sends file to client 2
func TestConnect(t *testing.T) {
	client, err := InitConfig()
	if err != nil {
		t.Error(err)
	}
	if err != nil {
		return
	}
	log.Debug("init")
	pubBlock, privBlock, err := cryptography.OpenKeys(client.KeyPath)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	client.pubKeyBlock = pubBlock
	client.privKey, err = cryptography.PemToKeys(privBlock)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer func() {
		err = client.Disconnect()
	}()

	err = client.Connect()
	if err != nil {
		t.Error(err)
		return
	}

	log.Debug("connect")

	if err != nil {
		t.Error(err)
		return
	}
	err = client.DoRequestP2P(client.conn, []byte("abcd"))
	if err != nil {
		log.Error(err)
	}
	log.Debug("holepunch")
}

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
//	if client.contactList != nil {
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
//	client.contactList = nil
//	err = client.ReadContactsFile()
//	if client.contactList == nil {
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
//	if client.removeContact([]byte("123456")) {
//		log.Debug("successfully removed")
//	} else {
//		log.Debug("could not find contact")
//	}
//	if client.removeContact([]byte("123556")) {
//		log.Debug("successfully removed")
//	} else {
//		log.Debug("could not find contact")
//	}
//	if client.removeContact([]byte("abcdef")) {
//		log.Debug("successfully removed")
//	} else {
//		log.Debug("could not find contact")
//	}
//	log.Debug(client.contactList)
//}
