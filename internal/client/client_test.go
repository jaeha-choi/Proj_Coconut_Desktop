package client

import (
	"github.com/jaeha-choi/Proj_Coconut_Utility/cryptography"
	"github.com/jaeha-choi/Proj_Coconut_Utility/log"
	"os"
	"testing"
	"time"
)

func initClient(keyN string, log *log.Logger) Client {

	client := InitConfig(log)
	//client.KeyPath = "/home/duncan/projects/Proj_Coconut_Desktop"
	client.ServerHost = "coconut-demo.jaeha.dev"
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

//func TestDoOpenHolePunch(t *testing.T) {
//	// *P2P SERVER*
//	// add code
//	// request pubkey
//	l := log.NewLogger(os.Stdout, log.DEBUG, "P2P SERVER")
//	client := initClient("key", l)
//	defer func() {
//		_ = client.Disconnect()
//	}()
//	err := client.Connect()
//
//	if err != nil {
//		t.Error(err)
//	}
//	var key string
//	key = "giapph/kXJ7PAHfMzWeE8hoqgQ0nirjjo0TAOElS598=" // robin
//	//key = "su+oF6panqRPm8cPyRJ9cAnlPFbEjzPgsIkaPbqNee4=" // jaeha
//	//key = "GoLvuVi0pf5tf4oqbRK1iex0aK56xjeMQR8vIykzS1U=" // duncan
//	err = client.DoRequestP2P([]byte(key))
//	if err != nil {
//		log.Error(err)
//	}
//	time.Sleep(10 * time.Second)
//
//}
//
//func TestDoOpenHolePunch2(t *testing.T) {
//	// *P2P CLIENT*
//	// get add code
//	l := log.NewLogger(os.Stdout, log.DEBUG, "P2P CLIENT")
//	client := initClient("keykey", l)
//	defer func() {
//		_ = client.Disconnect()
//	}()
//	_ = client.Connect()
//
//	client.addContact("jaeha", "choi", []byte("su+oF6panqRPm8cPyRJ9cAnlPFbEjzPgsIkaPbqNee4="), "server.pub")
//	client.addContact("robin", "seo", []byte("FBkHZ6e+q4yxaE9TsvPtFbE9HF1vpJP2MnWjvmWWiGI="), "server.pub")
//	client.addContact("duncan", "spani", []byte("GoLvuVi0pf5tf4oqbRK1iex0aK56xjeMQR8vIykzS1U="), "server.pub")
//	client.addContact("duncan2", "spani2", []byte("pGlM84wEUTwm9S5tmMEca5YvBcUyw26FdNJ75GkqwtE="), "server.pub")
//
//	time.Sleep(1 * time.Minute)
//}
//
//func TestDoOpenHolePunchLocalHost(t *testing.T) {
//	testDataPath := "../../testdata/"
//	l := log.NewLogger(os.Stdout, log.DEBUG, "SERVER ")
//	server := initClient("key", l)
//	l2 := log.NewLogger(os.Stdout, log.DEBUG, "CLIENT ")
//	client := initClient("keyJae", l2)
//	defer func() {
//		err := server.Disconnect()
//		if err != nil {
//			server.logger.Error(err)
//		}
//		err = client.Disconnect()
//		if err != nil {
//			client.logger.Error(err)
//		}
//	}()
//	go func() {
//		// *P2P CLIENT*
//		time.Sleep(1 * time.Second)
//
//		_ = client.Connect()
//		client.logger.Info(client.conn.LocalAddr())
//		err := client.DoGetAddCode()
//		if err != nil {
//			return
//		}
//		for {
//			if server.addCode != "" {
//				break
//			}
//			time.Sleep(500 * time.Millisecond)
//		}
//		client.logger.Info("server add code: ", server.addCode)
//		time.Sleep(5 * time.Second)
//		client.logger.Info("requesting server pubkey")
//		err = client.DoRequestPubKey(server.addCode, "server.key")
//		if err != nil {
//			client.logger.Error(err)
//		}
//		client.logger.Info("end request Pubkey")
//
//		client.addContact("jaeha", "choi", []byte(client.peerKey), "server.key")
//
//		client.logger.Info("end")
//		time.Sleep(1 * time.Minute)
//	}()
//
//	time.Sleep(1 * time.Second)
//
//	go func() {
//		// *P2P SERVER*
//
//		err := server.Connect()
//		server.logger.Info(server.conn.LocalAddr())
//		if err != nil {
//			t.Error(err)
//		}
//		err = server.DoGetAddCode()
//		if err != nil {
//			server.logger.Error(err)
//		}
//		for {
//			if client.addCode != "" {
//				break
//			}
//			time.Sleep(500 * time.Millisecond)
//		}
//		server.logger.Info("client add code: ", client.addCode)
//		err = server.DoRequestPubKey(client.addCode, "client.key")
//		if err != nil {
//			server.logger.Error(err)
//		}
//		server.logger.Debug("end pubkey")
//		time.Sleep(10 * time.Second)
//		////key := "giapph/kXJ7PAHfMzWeE8hoqgQ0nirjjo0TAOElS598=" // robin
//		////key := "haGoLvuVi0pf5tf4oqbRK1iex0aK56xjeMQR8vIykzS1U=" // duncan
//		key := "su+oF6panqRPm8cPyRJ9cAnlPFbEjzPgsIkaPbqNee4=" // jaeha
//		server.addContact("jaeha", "choi", []byte("su+oF6panqRPm8cPyRJ9cAnlPFbEjzPgsIkaPbqNee4="), "client.key")
//		server.logger.Debug("Starting request p2p")
//		err = server.DoRequestP2P([]byte(key))
//		if err != nil {
//			log.Error(err)
//		}
//		server.logger.Debug("end requestp2p")
//		time.Sleep(1 * time.Second)
//		server.logger.Debug("Starting send file")
//		err = server.DoSendFile(testDataPath + "1000000 Sales Records(119MB).csv")
//		if err != nil {
//			return
//		}
//		server.logger.Debug("end sendFile")
//		server.logger.Info("end")
//		time.Sleep(1 * time.Minute)
//	}()
//
//	time.Sleep(1 * time.Minute)
//
//}

// Remote testing to send file to peer
func TestClient_DoSendFile(t *testing.T) {
	testDataPath := "../../testdata/"
	logger := log.NewLogger(os.Stdout, log.DEBUG, "SERVER ")
	client := initClient("key", logger)

	_ = client.Connect()
	defer func() {
		err := client.Disconnect()
		if err != nil {
			client.logger.Error(err)
		}
	}()

	err := client.DoGetAddCode()
	if err != nil {
		return
	}
	client.logger.Info("My AddCode: ", client.addCode)
	var peerCode string
	client.logger.Debug(peerCode)

	err = client.DoRequestPubKey(peerCode, "peer.key")
	if err != nil {
		client.logger.Error(err)
	}

	client.addContact("first", "last", []byte(client.peerKey), "peer.key")

	time.Sleep(3 * time.Second)

	err = client.DoRequestP2P([]byte(client.peerKey))
	if err != nil {
		log.Error(err)
	}

	time.Sleep(1 * time.Second)

	err = client.DoSendFile(testDataPath + "1000000 Sales Records(119MB).csv")
	if err != nil {
		return
	}

}

// Remote testing to get file from peer
func TestPeer_HandleGetFile(t *testing.T) {
	logger := log.NewLogger(os.Stdout, log.DEBUG, "CLIENT ")
	client := initClient("keyJae", logger)
	_ = client.Connect()
	defer func() {
		err := client.Disconnect()
		if err != nil {
			client.logger.Error(err)
		}
	}()
	err := client.DoGetAddCode()
	if err != nil {
		return
	}
	var peerCode string
	client.logger.Debug(peerCode)

	err = client.DoRequestPubKey(peerCode, "peer.key")
	if err != nil {
		client.logger.Error(err)
	}

	client.addContact("first", "last", []byte(client.peerKey), "peer.key")
	// command handler running in background
	time.Sleep(1 * time.Minute)

}

//func TestKeys(t *testing.T) {
//l := log.NewLogger(os.Stdout, log.DEBUG, "SERVER ")
//server := initClient("key", l)
//l2 := log.NewLogger(os.Stdout, log.DEBUG, "CLIENT ")
////client := initClient("keyJae", l2)
//val := []byte{65,181,82,66,176,14,144,69,202,73,50,42,50,7,82,160,199,94,117,118,126,38,203,220,122,141,83,153,135,18,48,198,22,205,31,74,108,253,124,37,10,58,210,225,115,208,157,19,143,64,166,0,192,127,11,156,231,107,149,235,68,222,165,52,91,60,223,220,10,216,79,240,36,175,155,25,81,218,173,186,146,158,176,82,161,250,250,18,42,135,14,242,165,46,7,115,185,26,246,93,98,164,218,19,135,3,215,126,220,89,12,37,200,193,126,71,252,44,18,55,166,236,117,169,208,255,148,113,38,88,74,247,17,166,238,245,169,167,74,164,49,185,162,244,9,189,138,86,83,117,226,1,75,101,9,222,87,25,7,43,44,31,183,226,57,230,128,66,107,252,174,45,112,215,181,147,58,92,157,226,179,217,69,211,35,59,100,192,217,40,92,181,219,107,20,49,86,246,90,75,119,71,210,70,204,13,196,56,107,82,122,58,50,15,183,172,149,1,52,232,122,69,37,221,98,55,202,165,80,216,7,232,107,42,241,211,2,202,56,128,147,239,232,228,139,123,82,137,21,166,95,98,218,209,46,62,224,99,60,35,207,6,10,186,122,160,120,251,232,50,88,237,56,22,77,15,82,236,120,247,195,101,85,43,68,203,3,227,186,169,219,52,161,194,153,190,236,75,135,34,221,116,58,106,121,234,161,244,189,43,133,214,59,89,219,74,50,246,222,53,223,143,206,37,205,162,236,109,135,10,207,171,72,35,103,211,84,254,248,174,138,86,98,252,223,132,48,115,7,143,52,243,45,88,107,101,173,122,145,59,130,239,249,7,76,115,17,118,228,109,156,233,185,69,160,99,164,11,38,56,104,181,186,203,79,13,89,173,199,153,8,20,227,174,126,202,223,80,100,98,227,223,10,87,245,216,18,10,174,145,1,52,248,66,72,52,48,246,179,207,203,120,91,253,221,244,105,155,57,203,121,126,168,84,142,250,219,244,168,217,243,57,110,82,41,156,254,223,189,55,72,17,223,249,9,118,77,130,233,167,88,249,5,213,173,194,119,101,137,168,61,82,28,149,68,22,76,124,35,45,75,110,60,180,168,106,52,89,39,28,165,28,69,228,247,164,50,15,226,233,87,91,118,222,47,127,227,165,229,126,48,185}
//msg ,_,  _ := util.ReadMessageUDP(nil, val)
//log.Debug(msg.ErrorCode, msg.CommandCode, msg.Data)
//var key []byte = []byte("qdcskxcmf55h89g0axwh029k3up0j5zz")
//dataEncrypted, dataSignature, err := cryptography.EncryptSignMsg(key, &client.privKey.PublicKey, server.privKey)
//if err != nil {
//	log.Debug(err)
//	log.Error("Error in EncryptSignMsg")
//	t.Error(err)
//	return
//}
//
//key2, err := cryptography.DecryptVerifyMsg(dataEncrypted, dataSignature,  &server.privKey.PublicKey, client.privKey)
//if err != nil {
//	log.Debug(err)
//	log.Error("Error in DecryptVerifyMsg")
//	t.Error(err)
//	return
//}
//log.Debug(string(key2))
//}

//func TestDeadline(t *testing.T){
//	addr := "127.0.0.1"
//	l, _ := net.ResolveUDPAddr("udp", addr)
//	list, _ := net.ListenUDP("udp", l)
//	err := list.SetReadDeadline(time.Now().Add(5 * time.Second))
//	if err != nil {
//		log.Error(err)
//	}
//	buffer := make([]byte,1)
//	_, _, err = list.ReadFromUDP(buffer)
//	if err != nil{
//
//	}
//}

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
