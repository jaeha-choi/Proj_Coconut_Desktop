package client

import (
	"bytes"
	"crypto/rsa"
	"crypto/tls"
	"encoding/gob"
	"encoding/pem"
	"github.com/jaeha-choi/Proj_Coconut_Utility/common"
	"github.com/jaeha-choi/Proj_Coconut_Utility/cryptography"
	"github.com/jaeha-choi/Proj_Coconut_Utility/log"
	"github.com/jaeha-choi/Proj_Coconut_Utility/util"
	"gopkg.in/yaml.v3"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
)

const (
	keyPath  = "./"
	dataPath = "./data/"
)

type Client struct {
	ServerHost  string `yaml:"server_host"`
	ServerPort  uint16 `yaml:"server_port"`
	LocalPort   uint16 `yaml:"local_port"`
	KeyPath     string `yaml:"key_path"`
	DataPath    string `yaml:"data_path"`
	tlsConfig   *tls.Config
	privKey     *rsa.PrivateKey
	pubKeyBlock *pem.Block
	conn        net.Conn
	localAddr   net.Addr
	addCode     string
	contactList []Contact
}

type Contact struct {
	FirstName  string
	LastName   string
	PubKeyHash []byte
	PubKey     *pem.Block
}

// InitConfig initializes Client struct.
func InitConfig() (client *Client, err error) {
	client = &Client{
		ServerHost:  "127.0.0.1",
		ServerPort:  9129,
		LocalPort:   10378,
		tlsConfig:   &tls.Config{InsecureSkipVerify: true}, // TODO: Update after using trusted cert
		privKey:     nil,
		pubKeyBlock: nil,
		conn:        nil,
		localAddr:   nil,
		addCode:     "",
		KeyPath:     keyPath,
		DataPath:    dataPath,
	}
	return client, nil
}

// ReadConfig reads a config from a yaml file
func ReadConfig(fileName string) (client *Client, err error) {
	client, err = InitConfig()
	if err != nil {
		log.Debug(err)
		return nil, err
	}
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Debug(err)
		return nil, err
	}
	err = yaml.Unmarshal(file, &client)
	if err != nil {
		log.Debug(err)
		log.Error("Error while parsing config.yml")
		return nil, err
	}
	return client, nil
}

func (client *Client) HandleGetPubKey(conn net.Conn) (err error) {
	if _, err = util.WriteBytes(conn, client.pubKeyBlock.Bytes); err != nil {
		log.Debug(err)
		log.Error("Error while sending public key")
		return err
	}
	// HandleGetPubkey TODO: Write error code to conn?
	return nil
}

func (client *Client) doInit() (err error) {
	pubKeyHash := cryptography.PemToSha256(client.pubKeyBlock)
	if _, err = util.WriteBytes(client.conn, pubKeyHash); err != nil {
		log.Debug(err)
		log.Error("Error while init command")
		return err
	}

	return client.getResult(client.conn)
}

func (client *Client) doQuit() (err error) {
	if _, err = util.WriteString(client.conn, common.Quit.String()); err != nil {
		log.Debug(err)
		log.Error("Error while quit command")
		return err
	}

	return client.getResult(client.conn)
}

func (client *Client) DoGetAddCode() (err error) {
	if _, err = util.WriteString(client.conn, common.GetAddCode.String()); err != nil {
		return err
	}
	addCode, err := util.ReadBytes(client.conn)
	if err != nil {
		log.Debug(err)
		return err
	}
	client.addCode = string(addCode)

	return client.getResult(client.conn)
}

func (client *Client) DoRemoveAddCode() (err error) {
	if _, err = util.WriteString(client.conn, common.RemoveAddCode.String()); err != nil {
		return err
	}
	if _, err = util.WriteString(client.conn, client.addCode); err != nil {
		return err
	}

	return client.getResult(client.conn)
}

func (client *Client) DoRequestRelay(rxPubKeyHash string) (err error) {
	if _, err = util.WriteString(client.conn, common.RequestRelay.String()); err != nil {
		return err
	}
	if _, err = util.WriteString(client.conn, rxPubKeyHash); err != nil {
		return err
	}
	// DoRequestRelay TODO: Finish implementing

	return client.getResult(client.conn)
}

func (client *Client) DoRequestPubKey(rxAddCodeStr string, fileName string) (err error) {
	if _, err = util.WriteString(client.conn, common.RequestPubKey.String()); err != nil {
		return err
	}
	if _, err = util.WriteString(client.conn, rxAddCodeStr); err != nil {
		return err
	}
	rxPubKeyBytes, err := util.ReadBytes(client.conn)
	if err != nil {
		log.Debug(err)
		return err
	}
	if err = cryptography.BytesToPemFile(rxPubKeyBytes, fileName); err != nil {
		return err
	}

	return client.getResult(client.conn)
}

func (client *Client) getResult(conn net.Conn) (err error) {
	_, err = util.ReadBytes(conn)
	return err
}

func (client *Client) Connect() (err error) {
	if client.conn != nil {
		// Client already established active connection
		return common.ExistingConnError
	}
	log.Debug("Connecting...")
	client.conn, err = tls.Dial("tcp", client.ServerHost+":"+strconv.Itoa(int(client.ServerPort)), client.tlsConfig)
	if err != nil {
		log.Debug(err)
		log.Error("Error while connecting to the server")
		return err
	}

	log.Info(client.conn.LocalAddr().String())
	return client.doInit()
}

func (client *Client) Disconnect() (err error) {
	if client.conn == nil {
		return nil
	}
	log.Debug("Disconnecting...")
	if err = client.doQuit(); err != nil {
		log.Debug(err)
		log.Error("Task is not complete")
		return common.TaskNotCompleteError
	}
	if err = client.conn.Close(); err != nil {
		log.Debug(err)
		log.Error("Error while disconnecting from the server")
		return err
	}
	client.conn = nil
	return nil
}

func (client *Client) DoRequestP2P(conn net.Conn, pkHash []byte) (err error) {
	_, err = util.WriteString(conn, common.RequestPTP.String())
	if err != nil {
		log.Error("Error writing to server")
		return err
	}
	command, err := util.ReadString(conn)
	if command != "GKEY" || err != nil {
		return common.UnknownCommandError
	}

	_, err = util.WriteBytes(conn, pkHash)
	if err != nil {
		return err
	}
	peerLocalAddr, err := util.ReadString(conn)
	peerPublicAddr, err := util.ReadString(conn)
	err = DoOpenHolePunch(peerLocalAddr, peerPublicAddr)
	return err
}

func DoOpenHolePunch(addr1 string, addr2 string) (err error) {
	log.Info("Local Addr: ", addr1, "Public Addr: ", addr2)

	return err
}

// DoSendLocalIP sends local ip to conn
// returns error is applicable
func (client *Client) DoSendLocalIP() (err error) {
	client.localAddr = client.conn.LocalAddr()
	if _, err = util.WriteString(client.conn, client.localAddr.String()); err != nil {
		return err
	}
	return client.getResult(client.conn)
}

// ReadContactsFile read the contents of contacts.gob into client.contactList
// returns error if applicable
func (client *Client) ReadContactsFile() (err error) {
	var contactsList []Contact

	file, err := os.OpenFile(filepath.Join(client.DataPath, "contacts.gob"), os.O_RDONLY|os.O_CREATE, 0666)

	if err != nil {
		log.Error("Error opening file: ", err)
		return err
	}

	defer func() {
		err = file.Close()
		if err != nil {
			log.Error("Error closing file: ", err)
		}
	}()

	dataDecode := gob.NewDecoder(file)
	err = dataDecode.Decode(&contactsList)
	if err == io.EOF {
		client.contactList = nil
		return nil
	}
	if err != nil {
		log.Error("Error decoding file: ", err)
		return err
	}
	client.contactList = contactsList
	return err
}

// WriteContactsFile write contents of contacts array into contacts.gob file
// returns error if error generated
func (client *Client) WriteContactsFile() (err error) {
	file, err := os.OpenFile(filepath.Join(client.DataPath, "contacts.gob"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Error("Error opening file: ", err)
		return err
	}

	defer func() {
		err = file.Close()
		if err != nil {
			log.Error("Error closing file: ", err)
		}
	}()

	dataEncoder := gob.NewEncoder(file)
	return dataEncoder.Encode(client.contactList)
}

// Contacts TODO change client.contactList from slice to map,  key : PubKeyHash
// addContact initializes new contact struct
// returns true if contact added or already in list, false otherwise
func (client *Client) addContact(fname string, lname string, pkhash []byte, pubkey *pem.Block) (inserted bool) {
	// check if contact already in list
	for i := range client.contactList {
		if bytes.Compare(client.contactList[i].PubKeyHash, pkhash) == 0 {
			return true
		}
	}
	// create contact if not in list
	contact := Contact{
		fname,
		lname,
		pkhash,
		pubkey,
	}
	client.contactList = append(client.contactList, contact)
	return true
}

// findContact returns bool and contact if pubkey hash is found in contacts list
// returns false and empty contact if not found
func (client *Client) findContact(pkhash []byte) (b bool, c Contact) {
	for i := range client.contactList {
		if bytes.Compare(client.contactList[i].PubKeyHash, pkhash) == 0 {
			return true, client.contactList[i]
		}
	}
	return false, Contact{}
}

// removeContact removes contact with specified pubkey hash
// returns true if found and removed, false if not found
func (client *Client) removeContact(pkhash []byte) (b bool) {
	for i := range client.contactList {
		if bytes.Compare(client.contactList[i].PubKeyHash, pkhash) == 0 {
			client.contactList = removeContactHelper(client.contactList, i)
			return true
		}
	}
	return false
}
func removeContactHelper(l []Contact, i int) []Contact {
	l[i] = l[len(l)-1]
	return l[:len(l)-1]
}
