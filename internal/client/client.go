package client

import (
	"crypto/rsa"
	"crypto/tls"
	"encoding/pem"
	"errors"
	"github.com/jaeha-choi/Proj_Coconut_Utility/commands"
	"github.com/jaeha-choi/Proj_Coconut_Utility/cryptography"
	"github.com/jaeha-choi/Proj_Coconut_Utility/log"
	"github.com/jaeha-choi/Proj_Coconut_Utility/util"
	"net"
	"os"
	"strconv"
)

const (
	keyPath = "./"
)

type Client struct {
	ServerIp    string `yaml:"server_ip"`
	ServerPort  uint16 `yaml:"server_port"`
	tlsConfig   *tls.Config
	privKey     *rsa.PrivateKey
	pubKeyBlock *pem.Block
	conn        net.Conn
	addCode     string
}

var ExistingConnError = errors.New("existing connection present in client struct")

var UnexpectedError = errors.New("unexpected command returned")

var ReceiverNotFound = errors.New("receiver was not found")

func init() {
	log.Init(os.Stdout, log.DEBUG)
}

func NewClient() (client *Client, err error) {
	// Open RSA Keys
	pubBlock, privBlock, err := cryptography.OpenKeys(keyPath)
	if err != nil {
		log.Debug(err)
		return nil, err
	}
	privK, err := cryptography.PemToKeys(privBlock)
	if err != nil {
		log.Debug(err)
		return nil, err
	}

	client = &Client{
		ServerIp:    "127.0.0.1",
		ServerPort:  9129,
		tlsConfig:   &tls.Config{InsecureSkipVerify: true},
		privKey:     privK,
		pubKeyBlock: pubBlock,
		conn:        nil,
	}
	return client, nil
}

func (client *Client) HandleRequestPubKey() {

}

func (client *Client) DoInit() (isComplete bool, err error) {
	if client.conn != nil {
		// Client already established active connection
		return false, ExistingConnError
	}
	pubKeyHash := cryptography.PemToSha256(client.pubKeyBlock)
	if _, err = util.WriteBytes(client.conn, pubKeyHash); err != nil {
		log.Debug(err)
		return false, err
	}
	response, err := client.getResult(client.conn)
	if err != nil || !response {
		return false, err
	}
	return client.getResult(client.conn)
}

func (client *Client) DoGetAddCode() (isComplete bool, err error) {
	if _, err := util.WriteString(client.conn, commands.GetAddCode); err != nil {
		return false, err
	}
	addCode, err := util.ReadBytes(client.conn)
	if err != nil {
		return false, err
	}
	client.addCode = string(addCode)
	if isComplete, err = client.getResult(client.conn); err != nil || !isComplete {
		return false, err
	}
	return client.getResult(client.conn)
}

func (client *Client) DoRemoveAddCode() (isComplete bool, err error) {
	if _, err := util.WriteString(client.conn, commands.RemoveAddCode); err != nil {
		return false, err
	}
	if _, err := util.WriteString(client.conn, client.addCode); err != nil {
		return false, err
	}
	return client.getResult(client.conn)
}

func (client *Client) DoRequestRelay(rxPubKeyHash string) (isComplete bool, err error) {
	if _, err := util.WriteString(client.conn, commands.RequestRelay); err != nil {
		return false, err
	}
	if _, err := util.WriteString(client.conn, rxPubKeyHash); err != nil {
		return false, err
	}
	isRxFound, err := client.getResult(client.conn)
	if err != nil {
		return false, err
	}
	if !isRxFound {
		return false, ReceiverNotFound
	}
	// TODO: Finish implementing
	return client.getResult(client.conn)
}

func (client *Client) DoGetPubKey(rxAddCodeStr string, fileName string) (isComplete bool, err error) {
	if _, err := util.WriteString(client.conn, commands.GetPubKey); err != nil {
		return false, err
	}
	if _, err := util.WriteString(client.conn, rxAddCodeStr); err != nil {
		return false, err
	}
	rxPubKeyBytes, err := util.ReadBytes(client.conn)
	if err != nil {
		return false, err
	}
	if err = cryptography.BytesToPemFile(rxPubKeyBytes, fileName); err != nil {
		return false, err
	}
	return client.getResult(client.conn)
}

func (client *Client) getResult(conn net.Conn) (isAffirmation bool, err error) {
	bytes, err := util.ReadString(conn)
	if err != nil {
		return false, err
	}
	switch bytes {
	case commands.Affirmation:
		return true, nil
	case commands.Negation:
		return false, nil
	default:
		return false, UnexpectedError
	}
}

func (client *Client) Connect() (err error) {
	dial, err := tls.Dial("tcp", client.ServerIp+":"+strconv.Itoa(int(client.ServerPort)), client.tlsConfig)
	if err != nil {
		log.Debug(err)
		log.Error("Error while connecting to the server")
		return err
	}
	client.conn = dial
	return nil
}
