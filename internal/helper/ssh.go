package helper

import (
	"bytes"
	"context"
	"github.com/openinfradev/tks-api/pkg/log"
	"io"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

type safeBuffer struct {
	buffer bytes.Buffer
	mu     sync.Mutex
}

func (w *safeBuffer) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buffer.Write(p)
}
func (w *safeBuffer) Bytes() []byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buffer.Bytes()
}
func (w *safeBuffer) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buffer.Reset()
}

type LogicSshWsSession struct {
	stdinPipe       io.WriteCloser
	comboOutput     *safeBuffer
	logBuff         *safeBuffer
	inputFilterBuff *safeBuffer
	session         *ssh.Session
	wsConn          *websocket.Conn
	isAdmin         bool
	IsFlagged       bool
}

func NewLogicSshWsSession(cols, rows int, isAdmin bool, sshClient *ssh.Client, wsConn *websocket.Conn) (*LogicSshWsSession, error) {
	sshSession, err := sshClient.NewSession()
	if err != nil {
		return nil, err
	}

	stdinP, err := sshSession.StdinPipe()
	if err != nil {
		return nil, err
	}

	comboWriter := new(safeBuffer)
	logBuf := new(safeBuffer)
	inputBuf := new(safeBuffer)
	sshSession.Stdout = comboWriter
	sshSession.Stderr = comboWriter

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // disable echo
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}
	// Request pseudo terminal
	if err := sshSession.RequestPty("xterm", rows, cols, modes); err != nil {
		return nil, err
	}
	// Start remote shell
	if err := sshSession.Shell(); err != nil {
		return nil, err
	}
	return &LogicSshWsSession{
		stdinPipe:       stdinP,
		comboOutput:     comboWriter,
		logBuff:         logBuf,
		inputFilterBuff: inputBuf,
		session:         sshSession,
		wsConn:          wsConn,
		isAdmin:         isAdmin,
		IsFlagged:       false,
	}, nil
}

func (sws *LogicSshWsSession) Close() {
	if sws.session != nil {
		sws.session.Close()
	}
	if sws.logBuff != nil {
		sws.logBuff = nil
	}
	if sws.comboOutput != nil {
		sws.comboOutput = nil
	}
}
func (sws *LogicSshWsSession) Start(quitChan chan bool) {
	ctx := context.Background()
	go sws.receiveWsMsg(ctx, quitChan)
	go sws.sendComboOutput(ctx, quitChan)
}

func (sws *LogicSshWsSession) receiveWsMsg(ctx context.Context, exitCh chan bool) {
	wsConn := sws.wsConn
	//tells other go routine quit
	defer setQuit(exitCh)

	for {
		select {
		case <-exitCh:
			return
		default:
			//read websocket msg
			_, wsData, err := wsConn.ReadMessage()
			if err != nil {
				//log.Error(err)
				//log.Error("reading webSocket message failed")
				return
			}
			//cmd := bytes.NewBuffer(wsData).String()
			//log.Debug(cmd)
			//unmashal bytes into struct
			/*
					if err := json.Unmarshal(wsData, &msgObj); err != nil {
						log.Error(r.Context(),"unmarshal websocket message failed")
					}
				//handle xterm.js stdin
				decodeBytes, err := base64.StdEncoding.DecodeString(cmd)
				if err != nil {
					log.Error(r.Context(),"websock cmd string base64 decoding failed")
				}
			*/

			sws.sendWebsocketInputCommandToSshSessionStdinPipe(ctx, wsData)
		}
	}
}

// sendWebsocketInputCommandToSshSessionStdinPipe
func (sws *LogicSshWsSession) sendWebsocketInputCommandToSshSessionStdinPipe(ctx context.Context, cmdBytes []byte) {
	if _, err := sws.stdinPipe.Write(cmdBytes); err != nil {
		log.Error(ctx, "ws cmd bytes write to ssh.stdin pipe failed")
	}
}

func (sws *LogicSshWsSession) sendComboOutput(ctx context.Context, exitCh chan bool) {
	wsConn := sws.wsConn
	defer setQuit(exitCh)

	//every 120ms write combine output bytes into websocket response
	tick := time.NewTicker(time.Millisecond * time.Duration(60))
	//for range time.Tick(120 * time.Millisecond){}
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			if sws.comboOutput == nil {
				return
			}
			bs := sws.comboOutput.Bytes()
			if len(bs) > 0 {
				err := wsConn.WriteMessage(websocket.TextMessage, bs)
				if err != nil {
					log.Error(ctx, "ssh sending combo output to webSocket failed")
				}
				_, err = sws.logBuff.Write(bs)
				if err != nil {
					log.Error(ctx, "combo output to log buffer failed")
				}
				sws.comboOutput.buffer.Reset()
			}

		case <-exitCh:
			return
		}
	}
}

func (sws *LogicSshWsSession) Wait(quitChan chan bool) {
	if err := sws.session.Wait(); err != nil {
		//log.Error("ssh session wait failed")
		setQuit(quitChan)
	}
}

func (sws *LogicSshWsSession) LogString() string {
	return sws.logBuff.buffer.String()
}

func setQuit(ch chan bool) {
	ch <- true
}
