package main

import (
	"cloudshell/internal/log"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/creack/pty"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func handleXTermJS(w http.ResponseWriter, r *http.Request) {
	maxBufferSizeBytes := conf.GetInt("max-buffer-size-bytes")

	log.Debug("establishing connection uuid...")
	connectionUUID, err := uuid.NewUUID()
	if err != nil {
		message := fmt.Sprintf("failed to get a connection uuid: %s", err)
		log.Error(message)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(message))
		return
	}
	log.Infof("established connection uuid %s", connectionUUID.String())

	allowedHostnames := conf.GetStringSlice("allowed-hostnames")
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			requesterHostname := r.Host
			if strings.Index(requesterHostname, ":") != -1 {
				requesterHostname = strings.Split(requesterHostname, ":")[0]
			}
			for _, allowedHostname := range allowedHostnames {
				if requesterHostname == allowedHostname {
					return true
				}
			}
			log.Warnf("failed to find '%s' in the allowed hostnames", requesterHostname)
			return false
		},
		ReadBufferSize:  maxBufferSizeBytes,
		WriteBufferSize: maxBufferSizeBytes,
	}
	connection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Warnf("failed to upgrade connection: %s", err)
		return
	}

	terminal := conf.GetString("terminal-command")
	args := conf.GetStringSlice("terminal-args")
	log.Infof("starting new tty using command '%s' with arguments ['%s']...", terminal, strings.Join(args, "', '"))
	cmd := exec.Command(terminal, args...)
	cmd.Env = os.Environ()
	tty, err := pty.Start(cmd)
	if err != nil {
		message := fmt.Sprintf("failed to start tty: %s", err)
		log.Error(message)
		connection.WriteMessage(websocket.TextMessage, []byte(message))
		return
	}
	defer func() {
		cmd.Process.Kill()
		cmd.Process.Wait()
		tty.Close()
		connection.Close()
	}()

	var connectionClosed bool
	var waiter sync.WaitGroup
	waiter.Add(1)

	// this outputs from the pty
	go func() {
		for {
			buffer := make([]byte, 1024)
			readLength, err := tty.Read(buffer)
			if err != nil {
				message := fmt.Sprintf("failed to read from tty: %s", err)
				log.Warn(message)
				connection.WriteMessage(websocket.TextMessage, []byte("bye!"))
				waiter.Done()
				return
			}
			connection.WriteMessage(websocket.BinaryMessage, buffer[:readLength])
		}
	}()

	// this inputs into the pty
	go func() {
		for {
			// data processing
			messageType, reader, err := connection.NextReader()
			if err != nil {
				if !connectionClosed {
					message := fmt.Sprintf("failed to get next reader: %s", err)
					log.Warn(message)
				}
				return
			}
			dataBuffer := make([]byte, maxBufferSizeBytes)
			dataLength, err := reader.Read(dataBuffer)
			if err != nil {
				message := fmt.Sprintf("failed to get data type: %s", err)
				log.Warn(message)
				return
			}
			// debug
			switch messageType {
			case websocket.BinaryMessage:
				log.Tracef("received binary message (type: %v)", messageType)
				if dataBuffer[0] == 1 {
					var ttySize TTYSize
					resizeMessage := strings.Trim(string(dataBuffer), " \n\r\t\x00\x01")
					if err := json.Unmarshal([]byte(resizeMessage), &ttySize); err != nil {
						log.Warnf("failed to unmarshal received resize message '%s': %s", string(resizeMessage), err)
						continue
					}
					log.Infof("resizing tty to use %v rows and %v columns...", ttySize.Rows, ttySize.Cols)
					if err := pty.Setsize(tty, &pty.Winsize{
						Rows: ttySize.Rows,
						Cols: ttySize.Cols,
					}); err != nil {
						log.Warnf("failed to resize tty, error: %s", err)
					}
					continue
				}
			case websocket.TextMessage:
				log.Tracef("received text message (type: %v)", messageType)
			case websocket.CloseMessage:
				log.Tracef("received close message (type: %v)", messageType)
			case websocket.PingMessage:
				log.Tracef("received ping message (type: %v)", messageType)
			case websocket.PongMessage:
				log.Tracef("received ping message (type: %v)", messageType)
			}
			// get the key sequence
			var keySequence strings.Builder
			keySequenceLength := 0
			for _, keyCode := range dataBuffer {
				if keyCode == 0 {
					break
				}
				keySequence.WriteString(fmt.Sprintf(" %v", keyCode))
				keySequenceLength++
			}
			keySequenceString := strings.Trim(keySequence.String(), " ")
			log.Tracef("received key seq of %v bytes: %s",
				dataLength,
				keySequenceString,
			)
			switch true {
			case dataLength == -1: // invalid
				message := fmt.Sprintf("failed to get the correct number of bytes read, ignoring this")
				log.Warn(message)
			case dataLength > 0:
				dataToWrite := dataBuffer[:keySequenceLength]
				log.Debugf("writing %v bytes to tty...", len(dataToWrite))
				bytesWritten, err := tty.Write(dataToWrite)
				if err != nil {
					message := fmt.Sprintf("failed to write %v bytes to tty: %s", len(dataToWrite), err)
					log.Warn(message)
					continue
				}
				log.Debugf("%v bytes written to tty...", bytesWritten)
			default:

			}
		}
	}()

	waiter.Wait()
	connectionClosed = true
	log.Info("closing connection...")
}
