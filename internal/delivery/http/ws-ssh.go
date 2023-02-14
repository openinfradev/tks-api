package http

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/openinfradev/tks-common/pkg/log"
	"golang.org/x/crypto/ssh"

	"github.com/openinfradev/tks-api/internal/helper"
)

type Machine struct {
	Name     string `json:"name" gorm:"type:varchar(50);unique_index"`
	Host     string `json:"host" gorm:"type:varchar(50)"`
	Port     uint   `json:"port" gorm:"type:int(6)"`
	User     string `json:"user" gorm:"type:varchar(20)"`
	Password string `json:"password,omitempty"`
	Key      string `json:"key,omitempty"`
	Type     string `json:"type" gorm:"type:varchar(20)"`
}

var upGrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024 * 1024 * 10,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func publicKeyAuthFunc(keyPath string) ssh.AuthMethod {
	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		log.Error("ssh key file read failed", err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Error("ssh key signer failed", err)
	}
	return ssh.PublicKeys(signer)
}

func NewSshClient(h *Machine) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		Timeout:         time.Second * 5,
		User:            h.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	if h.Type == "password" {
		config.Auth = []ssh.AuthMethod{ssh.Password(h.Password)}
	} else {
		config.Auth = []ssh.AuthMethod{publicKeyAuthFunc(h.Key)}
	}
	addr := fmt.Sprintf("%s:%d", h.Host, h.Port)
	c, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (h *APIHandler) WsSsh(w http.ResponseWriter, r *http.Request) {
	wsConn, err := upGrader.Upgrade(w, r, nil)
	if err != nil {
		InternalServerError(w)
		return
	}
	defer wsConn.Close()

	client, err := NewSshClient(&Machine{
		User: "ubuntu",
		Host: "15.165.237.123",
		Port: 22,
		Type: "pem",
		Key:  "/Users/1110640/openinfradev/keys/tks-seoul.pem",
	})
	if err != nil {
		InternalServerError(w)
		return
	}
	defer client.Close()
	sws, err := helper.NewLogicSshWsSession(1000, 50, true, client, wsConn)
	if err != nil {
		InternalServerError(w)
		return
	}

	quitChan := make(chan bool, 3)
	sws.Start(quitChan)
	go sws.Wait(quitChan)

	<-quitChan

}
