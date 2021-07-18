package client

import (
	"crypto/tls"
	"github.com/jaeha-choi/Proj_Coconut_Utility/log"
	"strconv"
)

type Client struct {
	ServerIp   string `yaml:"server_ip"`
	ServerPort uint16 `yaml:"server_port"`
	tlsConfig  *tls.Config
}

func Connect() {
	client1 := &Client{
		ServerIp:   "127.0.0.1",
		ServerPort: 9129,
		tlsConfig:  &tls.Config{InsecureSkipVerify: true},
	}
	dial, err := tls.Dial("tcp", client1.ServerIp+":"+strconv.Itoa(int(client1.ServerPort)), client1.tlsConfig)
	if err != nil {
		log.Debug(err)
		log.Error("Error while connecting to the server")
		return
	}
	_, _ = dial.Write([]byte("hello"))
	_ = dial.Close()
}
