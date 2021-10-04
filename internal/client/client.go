package client

import (
	"crypto/rsa"
	"crypto/tls"
	"encoding/pem"
	"github.com/jaeha-choi/Proj_Coconut_Utility/common"
	"github.com/jaeha-choi/Proj_Coconut_Utility/cryptography"
	"github.com/jaeha-choi/Proj_Coconut_Utility/log"
	"github.com/jaeha-choi/Proj_Coconut_Utility/util"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"net"
	"strconv"
)

const (
	keyPath = "./"
)

type Client struct {
	ServerHost  string `yaml:"server_host"`
	ServerPort  uint16 `yaml:"server_port"`
	KeyPath     string `yaml:"key_path"`
	tlsConfig   *tls.Config
	privKey     *rsa.PrivateKey
	pubKeyBlock *pem.Block
	conn        net.Conn
	addCode     string
}

// InitConfig initializes Client struct.
func InitConfig() (client *Client, err error) {
	client = &Client{
		ServerHost:  "127.0.0.1",
		ServerPort:  9129,
		tlsConfig:   &tls.Config{InsecureSkipVerify: true}, // TODO: Update after using trusted cert
		privKey:     nil,
		pubKeyBlock: nil,
		conn:        nil,
		addCode:     "",
		KeyPath:     keyPath,
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
	// TODO: Write error code to conn?
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
	// TODO: Finish implementing

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
