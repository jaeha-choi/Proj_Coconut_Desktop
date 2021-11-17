package client

import (
	"github.com/jaeha-choi/Proj_Coconut_Utility/cryptography"
	"github.com/jaeha-choi/Proj_Coconut_Utility/log"
	"github.com/jaeha-choi/Proj_Coconut_Utility/util"
	"os"
	"testing"
	"time"
	//"time"
)

func initClient(keyN string, log *log.Logger) Client {

	client := InitConfig(log)
	//client.KeyPath = "/home/duncan/projects/Proj_Coconut_Desktop"
	//client.ServerHost = "coconut-demo.jaeha.dev"
	pubBlock, err := cryptography.OpenKeysAsBlock(client.KeyPath, keyN+".pub")
	if err != nil {
		log.Error(err)
	}
	privBlock, err := cryptography.OpenPrivKey(client.KeyPath, keyN+".priv")
	if err != nil {
		log.Error(err)
	}
	client.pubKeyBlock = pubBlock
	client.privKey = privBlock
	return *client
}

func TestDoOpenHolePunch(t *testing.T) {
	// *P2P SERVER*
	// add code
	// request pubkey
	l := log.NewLogger(os.Stdout, log.DEBUG, "P2P SERVER")
	client := initClient("key", l)
	defer func() {
		_ = client.Disconnect()
	}()
	err := client.Connect()

	if err != nil {
		t.Error(err)
	}
	var key string
	key = "giapph/kXJ7PAHfMzWeE8hoqgQ0nirjjo0TAOElS598=" // robin
	//key = "su+oF6panqRPm8cPyRJ9cAnlPFbEjzPgsIkaPbqNee4=" // jaeha
	//key = "GoLvuVi0pf5tf4oqbRK1iex0aK56xjeMQR8vIykzS1U=" // duncan
	err = client.DoRequestP2P([]byte(key))
	if err != nil {
		log.Error(err)
	}
	time.Sleep(10 * time.Second)

}

func TestDoOpenHolePunch2(t *testing.T) {
	// *P2P CLIENT*
	// get add code
	l := log.NewLogger(os.Stdout, log.DEBUG, "P2P CLIENT")
	client := initClient("keykey", l)
	defer func() {
		_ = client.Disconnect()
	}()
	_ = client.Connect()

	client.addContact("jaeha", "choi", []byte("su+oF6panqRPm8cPyRJ9cAnlPFbEjzPgsIkaPbqNee4="), "server.pub")
	client.addContact("robin", "seo", []byte("FBkHZ6e+q4yxaE9TsvPtFbE9HF1vpJP2MnWjvmWWiGI="), "server.pub")
	client.addContact("duncan", "spani", []byte("GoLvuVi0pf5tf4oqbRK1iex0aK56xjeMQR8vIykzS1U="), "server.pub")
	client.addContact("duncan2", "spani2", []byte("pGlM84wEUTwm9S5tmMEca5YvBcUyw26FdNJ75GkqwtE="), "server.pub")

	time.Sleep(1 * time.Minute)
	msg, _ := util.ReadMessage(client.peerConn)
	log.Debug(string(msg.Data))
}

func TestDoOpenHolePunchLocalHost(t *testing.T) {
	l := log.NewLogger(os.Stdout, log.DEBUG, "SERVER ")
	server := initClient("key", l)
	l2 := log.NewLogger(os.Stdout, log.DEBUG, "CLIENT ")
	client := initClient("keyJae", l2)
	defer func() {
		err := server.Disconnect()
		if err != nil {
			server.logger.Error(err)
		}
		err = client.Disconnect()
		if err != nil {
			client.logger.Error(err)
		}
	}()
	go func() {
		// *P2P CLIENT*
		time.Sleep(1 * time.Second)

		_ = client.Connect()
		client.logger.Info(client.conn.LocalAddr())
		err := client.DoGetAddCode()
		if err != nil {
			return
		}
		for {
			if server.addCode != "" {
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
		client.logger.Info("server add code: ", server.addCode)
		time.Sleep(5 * time.Second)
		client.logger.Info("requesting server pubkey")
		err = client.DoRequestPubKey(server.addCode, "server.key")
		if err != nil {
			client.logger.Error(err)
		}
		client.logger.Info("end request Pubkey")

		client.addContact("jaeha", "choi", []byte("su+oF6panqRPm8cPyRJ9cAnlPFbEjzPgsIkaPbqNee4="), "server.pub")
		client.addContact("robin", "seo", []byte("FBkHZ6e+q4yxaE9TsvPtFbE9HF1vpJP2MnWjvmWWiGI="), "server.pub")
		client.addContact("duncan", "spani", []byte("GoLvuVi0pf5tf4oqbRK1iex0aK56xjeMQR8vIykzS1U="), "server.key")
		client.addContact("duncan2", "spani2", []byte("haGoLvuVi0pf5tf4oqbRK1iex0aK56xjeMQR8vIykzS1U="), "server.pub")
		client.addContact("duncan", "spani", []byte("Wcrk//snVV+2hsNIGwVnrvsu4Txfj2YsbVVYVYTGxr0="), "server.key")

		client.logger.Info("end")
		time.Sleep(1 * time.Minute)
	}()

	time.Sleep(1 * time.Second)

	go func() {
		// *P2P SERVER*

		err := server.Connect()
		server.logger.Info(server.conn.LocalAddr())
		if err != nil {
			t.Error(err)
		}
		err = server.DoGetAddCode()
		if err != nil {
			server.logger.Error(err)
		}
		for {
			if client.addCode != "" {
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
		server.logger.Info("client add code: ", client.addCode)
		err = server.DoRequestPubKey(client.addCode, "client.key")
		if err != nil {
			server.logger.Error(err)
		}
		server.logger.Debug("end pubkey")
		time.Sleep(10 * time.Second)
		////key := "giapph/kXJ7PAHfMzWeE8hoqgQ0nirjjo0TAOElS598=" // robin
		////key := "haGoLvuVi0pf5tf4oqbRK1iex0aK56xjeMQR8vIykzS1U=" // duncan
		key := "su+oF6panqRPm8cPyRJ9cAnlPFbEjzPgsIkaPbqNee4=" // jaeha
		server.addContact("jaeha", "choi", []byte("su+oF6panqRPm8cPyRJ9cAnlPFbEjzPgsIkaPbqNee4="), "client.key")
		server.logger.Debug("Starting request p2p")
		err = server.DoRequestP2P([]byte(key))
		if err != nil {
			log.Error(err)
		}
		server.logger.Debug("end requestp2p")
		time.Sleep(1 * time.Second)
		server.logger.Debug("Starting send file")
		err = server.DoSendFile("/home/duncan/Downloads/abc.txt")
		if err != nil {
			return
		}
		server.logger.Debug("end sendFile")
		server.logger.Info("end")
		time.Sleep(1 * time.Minute)
	}()

	time.Sleep(1 * time.Minute)

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
