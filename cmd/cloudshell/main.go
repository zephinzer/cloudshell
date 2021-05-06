package main

import (
	"cloudshell/internal/log"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
	"unsafe"

	"github.com/creack/pty"
	"github.com/google/uuid"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

func main() {
	command := &cobra.Command{
		RunE: func(_ *cobra.Command, args []string) error {
			// debug stuff
			terminalCommand := conf.GetString("terminal-command")
			log.Infof("using '\033[1m%s\033[0m' as the terminal command", terminalCommand)
			terminalArguments := strings.Join(conf.GetStringSlice("terminal-args"), "\033[0m\", \"\033[1m")
			log.Infof("using [\"\033[1m%s\033[0m\"] as the terminal arguments", terminalArguments)
			maxBufferSizeBytes := conf.GetInt("max-buffer-size-bytes")
			log.Infof("using '\033[1m%v\033[0m' as the maxmimum buffer size in bytes", maxBufferSizeBytes)
			serverAddress := conf.GetString("server-addr")
			log.Infof("using address '\033[1m%s\033[0m' as the server address", serverAddress)
			serverPort := conf.GetInt("server-port")
			log.Infof("using port '\033[1m%v\033[0m' as the server port", serverPort)
			workingDirectory := conf.GetString("workdir")
			if !path.IsAbs(workingDirectory) {
				wd, err := os.Getwd()
				if err != nil {
					message := fmt.Sprintf("failed to get working directory: %s", err)
					log.Error(message)
					return errors.New(message)
				}
				workingDirectory = path.Join(wd, workingDirectory)
			}
			log.Infof("using '\033[1m%s\033[0m' as the working directory", workingDirectory)

			router := mux.NewRouter()

			// readiness probe endpoint
			router.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("ok"))
			})

			// liveness probe endpoint
			router.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("ok"))
			})

			// this is the endpoint for xterm.js to connect to
			router.HandleFunc("/xterm.js", handleXTermJS)

			// this is the endpoint for serving xterm.js assets
			depenenciesDirectory := path.Join(workingDirectory, "./node_modules")
			router.PathPrefix("/assets").Handler(http.StripPrefix("/assets", http.FileServer(http.Dir(depenenciesDirectory))))

			// this is the endpoint for the root path aka website
			publicAssetsDirectory := path.Join(workingDirectory, "./public")
			router.PathPrefix("/").Handler(http.FileServer(http.Dir(publicAssetsDirectory)))

			// listen
			listenOnAddress := fmt.Sprintf("%s:%v", serverAddress, serverPort)
			server := http.Server{
				Addr:    listenOnAddress,
				Handler: addLoggingMiddleware(router),
			}

			log.Infof("starting server on address: %s...", listenOnAddress)
			return server.ListenAndServe()
		},
	}
	conf.ApplyToCobra(command)
	command.Execute()
}

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

	upgrader := websocket.Upgrader{
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

	// this outputs from the pty
	go func() {
		for {
			buffer := make([]byte, 1024)
			readLength, err := tty.Read(buffer)
			if err != nil {
				message := fmt.Sprintf("failed to read from tty: %s", err)
				log.Warn(message)
				connection.WriteMessage(websocket.TextMessage, []byte(message))
				return
			}
			connection.WriteMessage(websocket.BinaryMessage, buffer[:readLength])
		}
	}()

	// this inputs into the pty
	for {
		// data processing
		messageType, reader, err := connection.NextReader()
		if err != nil {
			message := fmt.Sprintf("failed to get next reader: %s", err)
			log.Warn(message)
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
				if _, _, err := syscall.Syscall(
					syscall.SYS_IOCTL,
					tty.Fd(),
					syscall.TIOCSWINSZ,
					uintptr(unsafe.Pointer(&ttySize)),
				); err != 0 {
					log.Warnf("failed to resize tty, error number: %v", err)
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
		log.Tracef("received message of length %v: '%s', key seq: %s",
			dataLength,
			string(dataBuffer),
			keySequenceString,
		)
		switch true {
		case dataLength == -1: // invalid
			message := fmt.Sprintf("failed to get the correct number of bytes read, ignoring this")
			log.Warn(message)
		case dataLength > 0: // cli control keys
			log.Debug("sending message to pty...")
			bytesWritten, err := tty.Write(dataBuffer)
			if err != nil {
				message := fmt.Sprintf("failed to write %v bytes to tty: %s", bytesWritten, err)
				log.Warn(message)
				continue
			}
			log.Debugf("%v bytes written", bytesWritten)
		default:

		}
	}
}

type windowSize struct {
	Rows uint16 `json:"rows"`
	Cols uint16 `json:"cols"`
	X    uint16
	Y    uint16
}
