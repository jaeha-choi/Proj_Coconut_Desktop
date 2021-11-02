package client

import (
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
	"sync"
	"time"
)

const (
	// keyPath is a default path for asymmetric keys
	keyPath = "./"
	// dataPath is a default path for various data,
	// including UI interface, gob file that contains contact list, etc.
	dataPath = "./data/"
	//// contactsCapacity is the maximum number of contacts that a client can hold
	//contactsCapacity = 200
	// defaultServerPort is a default port number for central relay server
	defaultServerPort = 9129
	// defaultLocalPort is a default port number that will be opened for P2P connection
	defaultLocalPort = 10378
	// bufferSize is a default buffer size for each channels
	bufferSize = 10
)

// Client structure stores all necessary user data
type Client struct {
	// ServerHost is the central relay server's ip address
	ServerHost string `yaml:"server_host"`
	// ServerPort is the central relay server's port
	ServerPort uint16 `yaml:"server_port"`
	// LocalPort is a client port that will be opened for P2P connection,
	// in case hole punching fails
	LocalPort uint16 `yaml:"local_port"`
	// KeyPath is a path for asymmetric keys
	KeyPath string `yaml:"key_path"`
	// DataPath is a path for various data,
	// including UI interface, gob file that contains contact list, etc.
	DataPath string `yaml:"data_path"`
	// tlsConfig stores TLS configuration for connections between the central relay server
	tlsConfig *tls.Config
	// privKey stores the RSA private and public key of this client
	privKey *rsa.PrivateKey
	// pubKeyBlock stores the RSA public key of this client in PEM block format
	pubKeyBlock *pem.Block
	// conn is a connection to the central relay server
	conn net.Conn
	// peerConn is a p2p connection between other peer
	peerConn net.Conn
	// localAddr is a local address of this client
	localAddr net.Addr
	// addCode is the current Add Code associated with this client
	addCode string
	// contactMap stores the map of Contact structures. Uses public key hash string as a key
	contactMap map[string]*Contact
	// chanMap stores the map of channels. Uses command string as a key
	chanMap map[string]chan *util.Message
}

// Contact stores information about added contacts
type Contact struct {
	// FirstName and LastName is a name that can be set to distinguish added devices
	FirstName string
	LastName  string
	// PubKeyHash stores the SHA256 hash of added device's public key
	PubKeyHash []byte
	// PubKey stores the PEM formatted public key of added device's public key
	PubKey *pem.Block
}

// InitConfig initializes a default Client struct.
func InitConfig() (client *Client) {
	client = &Client{
		ServerHost:  "coconut-demo.jaeha.dev", // TODO: update this value after deploying the relay server
		ServerPort:  defaultServerPort,
		LocalPort:   defaultLocalPort,
		KeyPath:     keyPath,
		DataPath:    dataPath,
		tlsConfig:   &tls.Config{InsecureSkipVerify: true}, // TODO: Update after using trusted cert
		privKey:     nil,
		pubKeyBlock: nil,
		conn:        nil,
		peerConn:    nil,
		localAddr:   nil,
		addCode:     "",
		contactMap:  make(map[string]*Contact),
		chanMap:     make(map[string]chan *util.Message),
	}
	return client
}

// ReadConfig reads a config from a yaml file and override default settings
func ReadConfig(fileName string) (client *Client, err error) {
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Debug(err)
		return nil, err
	}

	client = InitConfig()
	err = yaml.Unmarshal(file, &client)
	if err != nil {
		log.Debug(err)
		log.Error("Error while parsing config.yml")
		return nil, err
	}
	return client, nil
}

// getResult is called at the end of each operation to check potential error
func (client *Client) getResult(command *common.Command) (err error) {
	msg := <-client.chanMap[command.String]
	delete(client.chanMap, command.String)
	errCode := common.ErrorCodes[msg.ErrorCode]
	if errCode != nil {
		return errCode
	}
	return nil
}

func (client *Client) commandHandler() {
	for {
		msg, err := util.ReadMessage(client.conn)
		if err == io.EOF {
			break
		} else if err != nil {
			log.Debug(err)
			break
		}
		command := common.CommandCodes[msg.CommandCode]
		if c, ok := client.chanMap[command.String]; ok {
			c <- msg
		} else if command == common.GetPubKey {
			err = client.handleGetPubKey()
		}
	}
}

// Connect connects this client to the relay server and initializes the connection by calling doInit
// Returns common.ExistingConnError if client is already connected
func (client *Client) Connect() (err error) {
	if client.conn != nil {
		// Client already established active connection
		return common.ExistingConnError
	}
	log.Debug("Connecting...")
	resolvedHost, err := net.ResolveIPAddr("ip", client.ServerHost)
	client.ServerHost = resolvedHost.String()
	client.conn, err = tls.Dial("tcp", client.ServerHost+":"+strconv.Itoa(int(client.ServerPort)), client.tlsConfig)
	if err != nil {
		log.Debug(err)
		log.Error("Error while connecting to the server")
		return err
	}
	client.localAddr = client.conn.LocalAddr()
	log.Debug("Connected")

	go client.commandHandler()

	return client.doInit()
}

// Disconnect disconnects this client from the server
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
	// Timer allows graceful shutdown for client.conn
	time.Sleep(1 * time.Second)
	if err = client.conn.Close(); err != nil {
		log.Debug(err)
		log.Error("Error while disconnecting from the server")
		return err
	}
	client.conn = nil
	log.Debug("Disconnected")
	return nil
}

// handleGetPubKey is called when the relay server requests this client's public key
func (client *Client) handleGetPubKey() (err error) {
	var command = common.GetAddCode
	client.chanMap[command.String] = make(chan *util.Message, bufferSize)
	defer delete(client.chanMap, command.String)

	if _, err = util.WriteMessage(client.conn, client.pubKeyBlock.Bytes, nil, command); err != nil {
		log.Debug(err)
		log.Error("Error while sending public key")
		return err
	}
	return nil
}

// doInit initializes the connection by sending this client's public key hash (SHA256) and
// private IP address to the relay server
func (client *Client) doInit() (err error) {
	var command = common.Init
	client.chanMap[command.String] = make(chan *util.Message, bufferSize)
	defer delete(client.chanMap, command.String)

	pubKeyHash := cryptography.PemToSha256(client.pubKeyBlock)
	if _, err = util.WriteMessage(client.conn, pubKeyHash, nil, command); err != nil {
		log.Debug(err)
		log.Error("Error while sending public key hash")
		return err
	}
	localAddress := client.localAddr.String()
	if _, err = util.WriteMessage(client.conn, []byte(localAddress), nil, command); err != nil {
		log.Debug(err)
		log.Error("Error while sending local ip address")
		return err
	}

	return client.getResult(command)
}

// doQuit signals the relay server to unregister this client
func (client *Client) doQuit() (err error) {
	var command = common.Quit
	client.chanMap[command.String] = make(chan *util.Message, bufferSize)
	defer delete(client.chanMap, command.String)

	if _, err = util.WriteMessage(client.conn, nil, nil, command); err != nil {
		log.Debug(err)
		log.Error("Error while quit command")
		return err
	}

	return client.getResult(command)
}

// DoGetAddCode signals the relay server to send the Add Code
// Returns common.NoAvailableAddCodeError if no Add Code is available
func (client *Client) DoGetAddCode() (err error) {
	var command = common.GetAddCode
	client.chanMap[command.String] = make(chan *util.Message, bufferSize)
	defer delete(client.chanMap, command.String)

	if _, err = util.WriteMessage(client.conn, nil, nil, command); err != nil {
		return err
	}
	msg := <-client.chanMap[command.String]
	client.addCode = string(msg.Data)

	return client.getResult(command)
}

// DoRemoveAddCode signals the relay server to dissociate the Add Code from this client
func (client *Client) DoRemoveAddCode() (err error) {
	var command = common.RemoveAddCode
	client.chanMap[command.String] = make(chan *util.Message, bufferSize)
	defer delete(client.chanMap, command.String)

	if _, err = util.WriteMessage(client.conn, nil, nil, command); err != nil {
		return err
	}
	if _, err = util.WriteMessage(client.conn, []byte(client.addCode), nil, command); err != nil {
		return err
	}

	return client.getResult(command)
}

// DoRequestRelay signals the relay server to relay files between this client and
// the client with matching rxPubKeyHash
// TODO: WIP
func (client *Client) DoRequestRelay(rxPubKeyHash string) (err error) {
	var command = common.RequestRelay
	client.chanMap[command.String] = make(chan *util.Message, bufferSize)
	defer delete(client.chanMap, command.String)

	if _, err = util.WriteMessage(client.conn, nil, nil, command); err != nil {
		return err
	}
	if _, err = util.WriteMessage(client.conn, []byte(rxPubKeyHash), nil, command); err != nil {
		return err
	}
	// Add file relay here

	return client.getResult(command)
}

// DoRequestPubKey signals the relay server to send public key associated with provided Add Code (rxAddCodeStr),
// then save it as fileName
// Returns common.ClientNotFoundError if no client is found
func (client *Client) DoRequestPubKey(rxAddCodeStr string, fileName string) (err error) {
	var command = common.RequestPubKey
	client.chanMap[command.String] = make(chan *util.Message, bufferSize)
	defer delete(client.chanMap, command.String)

	if _, err = util.WriteMessage(client.conn, nil, nil, command); err != nil {
		return err
	}
	if _, err = util.WriteMessage(client.conn, []byte(rxAddCodeStr), nil, command); err != nil {
		return err
	}

	// Get rxPubKeyBytes
	msg := <-client.chanMap[command.String]
	if err = cryptography.BytesToPemFile(msg.Data, fileName); err != nil {
		return err
	}

	return client.getResult(command)
}

// DoRequestP2P signals the relay server to ...
// TODO: WIP
func (client *Client) DoRequestP2P(pkHash []byte) (err error) {
	var command = common.RequestP2P
	client.chanMap[command.String] = make(chan *util.Message, bufferSize)

	if _, err = util.WriteMessage(client.conn, nil, nil, command); err != nil {
		log.Error("Error writing to server")
		return err
	}
	msg, err := util.ReadMessage(client.conn)
	if msg.CommandCode != common.GetP2PKey.Code || err != nil {
		return common.UnknownCommandError
	}

	_, err = util.WriteMessage(client.conn, pkHash, nil, common.GetP2PKey)
	if err != nil {
		return err
	}
	peerLocalAddr, err := util.ReadMessage(client.conn)
	peerPublicAddr, err := util.ReadMessage(client.conn)
	err = client.DoOpenHolePunch(string(peerLocalAddr.Data), string(peerPublicAddr.Data))
	return err
}

// DoOpenHolePunch Initiates the connection between client and peer
// returns connection stored in client.peerConn is connection made
// TODO: WIP
func (client *Client) DoOpenHolePunch(addr1 string, addr2 string) (err error) {
	log.Info("Local Addr: ", addr1, ", Public Addr: ", addr2)

	// create WaitGroup to halt processes until
	// both initPrivateAddr and initRemoteAddr complete
	var wg sync.WaitGroup

	wg.Add(1)
	go client.initP2PConn(&wg, addr1)

	wg.Add(1)
	go client.initP2PConn(&wg, addr2)

	// wait for goroutines to finish
	wg.Wait()

	if client.peerConn != nil {
		log.Info("Connection made to: ", client.peerConn.RemoteAddr())
	} else {
		log.Error("Unable to establish connection to peer")
		// TODO uncomment next line
		//return PeerUnavailableError
		return nil
	}

	return err
}

// initP2PConn initialize a connection with the provided.
// client.peerConn contains p2p connection if dialing was successful
// TODO: WIP
func (client *Client) initP2PConn(wg *sync.WaitGroup, addr string) {
	defer wg.Done()
	privBuffer := make([]byte, 1024)
	p2p, err := net.Dial("tcp", addr)
	if err != nil {
		log.Debug("Unable to connect: ", addr)
		return
	}
	log.Debug("Connection success: ", p2p.RemoteAddr())
	client.peerConn = p2p
	// TODO uncomment next line if `common.HolePunchPing.String()` exists
	//	_, _ = p2p.Write([]byte(common.HolePunchPing.String()))
	_, _ = p2p.Write([]byte("PING"))
	_, _ = p2p.Read(privBuffer)
	//i, _ := p2p.Read(privBuffer)
	//log.Debug(string(privBuffer[:i]))

}

// ReadContactsFile read the contents of contacts.gob into client.contactMap
func (client *Client) ReadContactsFile() (err error) {
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

	err = gob.NewDecoder(file).Decode(&client.contactMap)
	if err == io.EOF {
		//client.contactMap = nil
		return nil
	} else if err != nil {
		log.Debug(err)
		log.Error("Error decoding file: ", err)
		return err
	}
	return err
}

// WriteContactsFile write contents of contacts map into contacts.gob file
func (client *Client) WriteContactsFile() (err error) {
	file, err := os.OpenFile(filepath.Join(client.DataPath, "contacts.gob"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Debug(err)
		log.Error("Error opening file: ", err)
		return err
	}

	defer func() {
		err = file.Close()
		if err != nil {
			log.Error("Error closing file: ", err)
		}
	}()

	return gob.NewEncoder(file).Encode(client.contactMap)
}

// addContact initializes new contact struct
// Returns true if contact is added or already in list, false otherwise
func (client *Client) addContact(firstName string, lastName string, pkHash []byte, pubKey *pem.Block) (inserted bool) {
	pkHashStr := string(pkHash)
	// check if contact already in list
	if _, isFound := client.contactMap[pkHashStr]; isFound {
		return true
	}
	// create contact if not in map
	contact := Contact{
		firstName,
		lastName,
		pkHash,
		pubKey,
	}
	client.contactMap[pkHashStr] = &contact
	return true
}

// RemoveContact removes contact with specified public key hash
// Returns true if found and removed, false if not found
func (client *Client) RemoveContact(pkHash string) (b bool) {
	if _, exist := client.contactMap[pkHash]; exist {
		delete(client.contactMap, pkHash)
		return true
	}
	return false
}
