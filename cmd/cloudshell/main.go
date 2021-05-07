package main

import (
	"cloudshell/internal/log"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
)

var VersionInfo string

func main() {
	if VersionInfo == "" {
		VersionInfo = "dev"
	}
	command := &cobra.Command{
		Use:     "cloudshell",
		Short:   "Creates a web-based shell using xterm.js that links to an actual shell",
		Version: VersionInfo,
		RunE: func(_ *cobra.Command, args []string) error {
			// debug stuff
			terminalCommand := conf.GetString("terminal-command")
			log.Infof("using '\033[1m%s\033[0m' as the terminal command", terminalCommand)
			terminalArguments := strings.Join(conf.GetStringSlice("terminal-args"), "\033[0m\", \"\033[1m")
			log.Infof("using [\"\033[1m%s\033[0m\"] as the terminal arguments", terminalArguments)
			allowedHostnames := strings.Join(conf.GetStringSlice("allowed-hostnames"), "\033[0m\", \"\033[1m")
			log.Infof("using [\"\033[1m%s\033[0m\"] as the allowed hostnames", allowedHostnames)
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

			// this is the endpoint for xterm.js to connect to
			router.HandleFunc("/xterm.js", handleXTermJS)

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

			// version endpoint
			router.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(VersionInfo))
			})

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
