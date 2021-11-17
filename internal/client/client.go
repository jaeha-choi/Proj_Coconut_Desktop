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
	peerConn *net.UDPConn
	// peerAddr
	peerAddr *net.UDPAddr
	// peerKey is a p2p pubkey hash used to identify the client connected to
	peerKey string
	// localAddr is a local address of this client
	localAddr net.Addr
	// addCode is the current Add Code associated with this client
	addCode string
	// contactMap stores the map of Contact structures. Uses public key hash string as a key
	contactMap map[string]*Contact
	// chanMap stores the map of channels. Uses command string as a key
	chanMap map[string]chan *util.Message
	// logger
	logger *log.Logger
}

// Contact stores information about added contacts
type Contact struct {
	// FirstName and LastName is a name that can be set to distinguish added devices
	FirstName string
	LastName  string
	// PubKey stores the PEM formatted public key of added device's public key
	PubKeyFile string
}

// InitConfig initializes a default Client struct.
func InitConfig(log *log.Logger) (client *Client) {
	client = &Client{
		ServerHost:  "127.0.0.1", // TODO: update this value after deploying the relay server
		ServerPort:  defaultServerPort,
		LocalPort:   defaultLocalPort,
		KeyPath:     keyPath,
		DataPath:    dataPath,
		tlsConfig:   &tls.Config{InsecureSkipVerify: true}, // TODO: Update after using trusted cert
		privKey:     nil,
		pubKeyBlock: nil,
		conn:        nil,
		peerConn:    nil,
		peerKey:     "",
		localAddr:   nil,
		addCode:     "",
		contactMap:  make(map[string]*Contact),
		chanMap:     make(map[string]chan *util.Message),
		logger:      log,
	}
	return client
}

// ReadConfig reads a config from a yaml file and override default settings
func ReadConfig(fileName string, log *log.Logger) (client *Client, err error) {
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		client.logger.Debug(err)
		return nil, err
	}

	client = InitConfig(log)
	err = yaml.Unmarshal(file, &client)
	if err != nil {
		client.logger.Debug(err)
		client.logger.Error("Error while parsing config.yml")
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
			client.logger.Debug(err)
			break
		}
		command := common.CommandCodes[msg.CommandCode]
		// TODO: Error handling
		if c, ok := client.chanMap[command.String]; ok {
			c <- msg
		} else if command == common.RequestPubKey {
			go func() {
				err = client.handleGetPubKey()
			}()
		} else if command == common.HandleRequestP2P {
			go func() {
				err = client.HandleRequestP2P()
				go client.UDPCommandHandler()
			}()
		} else if command == common.File {
			go func() {
				// only handles getting files from server
				err = client.HandleGetFile()
			}()
		} else if command == common.Pause {
			// signals ending of previously sent pause command
			continue
		} else {
			err = common.UnknownCommandError
		}
		if err != nil {
			client.logger.Debug("command handler error: ", command)
			client.logger.Debug("command handler error: ", string(msg.Data))
			client.logger.Error(err)
		}
	}
}

func (client *Client) UDPCommandHandler() {
	client.logger.Debug("Entering UDP Command Handler")
	for {
		msg, err := util.ReadMessage(client.peerConn)
		client.logger.Info("UDP COMMAND HANDLER: ", msg.CommandCode, msg.ErrorCode, string(msg.Data))
		//client.logger.Debug(string(msg.Data))
		if err != nil {
			client.logger.Error(err)
		}
		command := common.CommandCodes[msg.CommandCode]
		// TODO: Error handling
		if c, ok := client.chanMap[command.String]; ok {
			c <- msg
		} else if command == common.Quit {
			log.Info("Returning to Command Handler")
			return
		} else if command == common.Init {
			client.logger.Debug("INIT COMMAND")
			continue
		} else if command == common.RequestPubKey {
			err = client.handleGetPubKey()
		} else if command == common.File {
			client.logger.Debug("FILE COMMAND")
			err = client.HandleGetFile()
		} else {
			err = common.UnknownCommandError
		}
		if err != nil {
			client.logger.Debug("UDP command handler error: ", command, string(msg.Data))
			client.logger.Error(err)
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
	client.logger.Debug("Connecting...")
	// TODO: Double check if resolving is necessary
	resolvedHost, err := net.ResolveIPAddr("ip", client.ServerHost)
	if err != nil {
		return err
	}
	client.ServerHost = resolvedHost.String()
	client.conn, err = tls.Dial("tcp", client.ServerHost+":"+strconv.Itoa(int(client.ServerPort)), client.tlsConfig)
	if err != nil {
		client.logger.Debug(err)
		client.logger.Error("Error while connecting to the server")
		return err
	}
	client.localAddr = client.conn.LocalAddr()
	client.logger.Debug("Connected")

	go client.commandHandler()

	return client.doInit()
}

// Disconnect disconnects this client from the server
func (client *Client) Disconnect() (err error) {
	if client.conn == nil {
		return nil
	}
	client.logger.Debug("Disconnecting...")
	if err = client.doQuit(); err != nil {
		client.logger.Debug(err)
		client.logger.Error("Task is not complete")
		return common.TaskNotCompleteError
	}
	// Timer allows graceful shutdown for client.conn
	time.Sleep(1 * time.Second)
	if err = client.conn.Close(); err != nil {
		client.logger.Debug(err)
		client.logger.Error("Error while disconnecting from the server")
		return err
	}
	client.conn = nil
	client.logger.Debug("Disconnected")
	return nil
}

// handleGetPubKey is called when the relay server requests this client's public key
func (client *Client) handleGetPubKey() (err error) {
	client.logger.Debug("Enter handleGetPubKey")
	var command = common.RequestPubKey
	client.chanMap[command.String] = make(chan *util.Message, bufferSize)
	defer func() {
		delete(client.chanMap, command.String)
	}()
	_, err = util.WriteMessage(client.conn, nil, nil, common.Pause)
	if err != nil {
		client.logger.Debug(err)
		return err
	}
	n, err := util.WriteMessage(client.conn, pem.EncodeToMemory(client.pubKeyBlock), nil, command)
	if err != nil {
		client.logger.Debug(err)
		client.logger.Error("Error while sending public key")
		return err
	}
	client.logger.Debug(n, " bytes written")
	return client.getResult(command)
}

// DoRequestPubKey signals the relay server to send public key associated with provided Add Code (rxAddCodeStr),
// then save it as fileName
// Returns common.ClientNotFoundError if no client is found
func (client *Client) DoRequestPubKey(rxAddCodeStr string, fileName string) (err error) {
	var command = common.RequestPubKey
	client.chanMap[command.String] = make(chan *util.Message, bufferSize)
	defer delete(client.chanMap, command.String)

	// write init command to server
	if _, err = util.WriteMessage(client.conn, nil, nil, command); err != nil {
		return err
	}
	// write rx add code to server
	if _, err = util.WriteMessage(client.conn, []byte(rxAddCodeStr), nil, command); err != nil {
		return err
	}
	msg := <-client.chanMap[command.String]
	client.peerKey = string(msg.Data)
	// Get rxPubKeyBytes
	msg = <-client.chanMap[command.String]
	// DO NOT CHANGE PERMISSION BITS
	f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		log.Debug(err)
		return err
	}
	defer f.Close()
	_, err = f.Write(msg.Data)
	if err != nil {
		return err
	}
	//client.contactMap[client.peerKey].PubKeyFile = fileName
	//if err = os.WriteFile(fileName, msg.Data, 0x777); err != nil {
	//	return err
	//}
	//if err = cryptography.BytesToPemFile(msg.Data, fileName); err != nil {
	//	return err
	//}

	//msg = <-client.chanMap[command.String]
	return client.getResult(command)
}

// doInit initializes the connection by sending this client's public key hash (SHA256) and
// private IP address to the relay server
func (client *Client) doInit() (err error) {
	var command = common.Init
	client.chanMap[command.String] = make(chan *util.Message, bufferSize)
	defer delete(client.chanMap, command.String)

	pubKeyHash := cryptography.PemToSha256(client.pubKeyBlock)
	if _, err = util.WriteMessage(client.conn, pubKeyHash, nil, command); err != nil {
		client.logger.Debug(err)
		client.logger.Error("Error while sending public key hash")
		return err
	}
	localAddress := client.localAddr.String()
	if _, err = util.WriteMessage(client.conn, []byte(localAddress), nil, command); err != nil {
		client.logger.Debug(err)
		client.logger.Error("Error while sending local ip address")
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
		client.logger.Debug(err)
		client.logger.Error("Error while quit command")
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

// DoSendFile encrypts and sends the file using the peers public key
func (client *Client) DoSendFile(fileName string) (err error) {
	client.logger.Debug("Sending file: ", fileName, " to ", client.peerAddr.String())
	var command = common.File
	client.chanMap[command.String] = make(chan *util.Message, bufferSize)
	defer delete(client.chanMap, command.String)
	// setup, encrypt
	// send init file command
	_, err = util.WriteMessageUDP(client.peerConn, client.peerAddr, []byte("a"), nil, command)
	_, err = util.WriteMessageUDP(client.peerConn, client.peerAddr, []byte("b"), nil, command)
	_, err = util.WriteMessageUDP(client.peerConn, client.peerAddr, []byte("c"), nil, command)
	_, err = util.WriteMessageUDP(client.peerConn, client.peerAddr, []byte("d"), nil, command)
	_, err = util.WriteMessageUDP(client.peerConn, client.peerAddr, []byte("e"), nil, command)
	if _, err = util.WriteMessageUDP(client.peerConn, client.peerAddr, []byte("f"), nil, command); err != nil {
		client.logger.Error("Error writing to peer")
		return err
	}
	chunk, err := cryptography.EncryptSetup(fileName)
	if err != nil {
		client.logger.Error("Error processing file: ", err)
		return err
	}

	//get RX pubkey from client stored in contact map
	//cryptography.OpenPubKey()
	if client.contactMap[client.peerKey] == nil {
		return common.ClientNotFoundError
	}
	// CONTACT MUST BE IN CONTACT MAP FOR THIS TO WORK
	pubKey, err := cryptography.OpenPubKey(keyPath, client.contactMap[client.peerKey].PubKeyFile)
	if err != nil {
		client.logger.Error(err)
		return common.PubKeyNotFoundError
	}
	//pubKey := client.contactMap[client.peerKey].PubKey

	//get TX privkey
	privKey := client.privKey

	// encrypt and send file
	err = chunk.EncryptUDP(client.peerConn, client.peerAddr, pubKey, privKey)
	if err != nil {
		client.logger.Error("Error encrypting file")
		return err
	}
	return err
}

// HandleGetFile receives and decrypts the specified file using private key
func (client *Client) HandleGetFile() (err error) {
	client.logger.Debug("Preparing to receive file")
	var command = common.File
	client.chanMap[command.String] = make(chan *util.Message, bufferSize)
	defer delete(client.chanMap, command.String)
	// setup, decrypt
	chunk, err := cryptography.DecryptSetup()
	if err != nil {
		client.logger.Error("Error setting up decryption")
		return err
	}

	//get RX pubkey from client stored in contact map
	//cryptography.OpenPubKey()
	pubKey, err := cryptography.OpenPubKey(keyPath, client.contactMap[client.peerKey].PubKeyFile)
	if err != nil {
		client.logger.Error(err)
		return common.PubKeyNotFoundError
	}
	//get TX privkey
	privKey := client.privKey

	// store file in downloads
	err = chunk.Decrypt(client.chanMap[command.String], pubKey, privKey)
	if err != nil {
		client.logger.Error("Error decrypting file")
		return err
	}
	client.logger.Debug("End receiving file")
	return err
}

func (client *Client) HandleRequestP2P() (err error) {
	// Init
	var command = common.HandleRequestP2P // TODO: Update to HandleRequestP2P
	client.chanMap[command.String] = make(chan *util.Message, bufferSize)
	defer delete(client.chanMap, command.String)

	// TODO: Fix rx -> server write msg issue
	//// 1b. Notify server that the rx is ready
	//if _, err := util.WriteMessage(client.conn, nil, nil, command); err != nil {
	//	return err
	//}

	// 2b. Receive tx public key hash from the server
	msg := <-client.chanMap[command.String]
	// Find tx using tx public key hash
	// TODO change "_" to "peerStruct" to retrieve pointer to structure
	_, ok := client.contactMap[string(msg.Data)]
	// 3b. Notify server that tx is found/not found
	if !ok {
		return common.ClientNotFoundError
	}
	client.peerKey = string(msg.Data)
	// TODO: Fix rx -> server write msg issue
	//if _, err := util.WriteMessage(client.conn, nil, nil, command); err != nil {
	//	client.logger.Debug("2 ", err)
	//	return err
	//}

	// Peer below allows access to full peer client structure
	// Possibly will be needed within ui
	//peer := *peerStruct

	// 4b. Receive tx localIP:localPort
	msg = <-client.chanMap[command.String]
	peerLocalAddr := msg.Data

	// 5b. Receive tx publicIP:publicPort
	msg = <-client.chanMap[command.String]
	peerRemoteAddr := msg.Data

	//// 6. Get result
	//if err = client.getResult(command); err != nil {
	//	return err
	//}

	//time.Sleep(500 * time.Millisecond)

	// Init hole punch
	return client.openHolePunch(string(peerLocalAddr), string(peerRemoteAddr))
}

// DoRequestP2P signals the relay server that a client wants to connect to another client
// TODO: WIP
func (client *Client) DoRequestP2P(pkHash []byte) (err error) {
	// Init
	var command = common.RequestP2P
	client.chanMap[command.String] = make(chan *util.Message, bufferSize)
	defer delete(client.chanMap, command.String)

	// 0. Write command
	if _, err = util.WriteMessage(client.conn, nil, nil, command); err != nil {
		client.logger.Error("Error writing to server")
		return err
	}

	// 1a. Read error code for finding tx client
	msg := <-client.chanMap[command.String]
	if msg.ErrorCode != 0 {
		return common.ErrorCodes[msg.ErrorCode]
	}

	// 2a. Write rx public key hash
	_, err = util.WriteMessage(client.conn, pkHash, nil, command)
	if err != nil {
		return err
	}
	client.peerKey = string(pkHash)

	// 3a. Read error code for finding rx client
	if msg := <-client.chanMap[command.String]; msg.ErrorCode != 0 {
		return common.ErrorCodes[msg.ErrorCode]
	}
	// 4a. Receive rx localIP:localPort to tx
	peerLocalAddr := <-client.chanMap[command.String]
	if peerLocalAddr.ErrorCode != 0 {
		return common.ErrorCodes[peerLocalAddr.ErrorCode]
	}

	// 5a. Receive rx publicIP:publicPort to tx
	peerPublicAddr := <-client.chanMap[command.String]
	if peerPublicAddr.ErrorCode != 0 {
		return common.ErrorCodes[peerPublicAddr.ErrorCode]
	}
	//for {
	//	msg := <-client.chanMap[command.String]
	//	client.logger.Debug(string(msg.Data))
	//
	//}
	//// 6. Get result
	//if err = client.getResult(command); err != nil {
	//	return err
	//}

	// Init hole punch
	return client.openHolePunch(string(peerLocalAddr.Data), string(peerPublicAddr.Data))
}

// openHolePunchClient Initiates the connection between client and peer
// stores UDP connection in client.peerConn and returns error is applicable
// TODO: remove local string
func (client *Client) openHolePunch(local string, remote string) (err error) {
	lAddrString := client.conn.LocalAddr().String()

	// Disconnect from relay server
	client.logger.Debug("HolePunch Initializing; Disconnecting From Server")
	if err := client.Disconnect(); err != nil {
		return err
	}

	client.logger.Info("Peer public Address: ", remote)

	// Resolve local address
	lAddr, err := net.ResolveUDPAddr("udp", lAddrString)

	if err != nil {
		client.logger.Debug(err)
		return err
	}

	client.peerConn, err = net.ListenUDP("udp", lAddr)
	client.peerAddr, err = net.ResolveUDPAddr("udp", remote)
	if err != nil {
		client.logger.Debug(err)
		return err
	}
	return err

}

// Reading
// **TX**
//go func() {
//	n, err := conn.Read(buffer)
//	if err != nil {
//		client.logger.Debug(err)
//	}
//	fmt.Println(string(buffer[:n]))
//	if err != nil {
//		return
//	}
//}()

// **RX**
//go func() {
//	client.logger.Debug("Reading file")
//	file, _ := os.OpenFile("checksum.txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
//	defer file.Close()
//	n, err := conn.Read(buffer)
//	if err != nil {
//		client.logger.Debug(err)
//	}
//	_, _ = file.Write(buffer[:n])
//	if err != nil {
//		return
//	}
//}()
// Writing
// **TX**
//file, _ := ioutil.ReadFile("checksum.txt")
//for {
//	_, err := conn.WriteTo(file,  addr)
//	if err != nil {
//		client.logger.Debug(err)
//		return err
//	}
//	time.Sleep(5 * time.Second)
//}
// **RX**
//for {
//	_, err := conn.WriteTo([]byte("hello"), addr)
//	if err != nil {
//		client.logger.Debug(err)
//		return err
//	}
//	time.Sleep(5 * time.Second)
//}

// ReadContactsFile read the contents of contacts.gob into client.contactMap
func (client *Client) ReadContactsFile() (err error) {
	file, err := os.OpenFile(filepath.Join(client.DataPath, "contacts.gob"), os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		client.logger.Error("Error opening file: ", err)
		return err
	}

	defer func() {
		err = file.Close()
		if err != nil {
			client.logger.Error("Error closing file: ", err)
		}
	}()

	err = gob.NewDecoder(file).Decode(&client.contactMap)
	if err == io.EOF {
		//client.contactMap = nil
		return nil
	} else if err != nil {
		client.logger.Debug(err)
		client.logger.Error("Error decoding file: ", err)
		return err
	}
	return err
}

// WriteContactsFile write contents of contacts map into contacts.gob file
func (client *Client) WriteContactsFile() (err error) {
	file, err := os.OpenFile(filepath.Join(client.DataPath, "contacts.gob"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		client.logger.Debug(err)
		client.logger.Error("Error opening file: ", err)
		return err
	}

	defer func() {
		err = file.Close()
		if err != nil {
			client.logger.Error("Error closing file: ", err)
		}
	}()

	return gob.NewEncoder(file).Encode(client.contactMap)
}

// addContact initializes new contact struct
// Returns true if contact is added or already in list, false otherwise
func (client *Client) addContact(firstName string, lastName string, pkHash []byte, fileName string) (inserted bool) {
	pkHashStr := string(pkHash)
	// check if contact already in list
	if _, isFound := client.contactMap[pkHashStr]; isFound {
		return true
	}
	// create contact if not in map
	contact := Contact{
		firstName,
		lastName,
		fileName,
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
