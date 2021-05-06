package main

import "github.com/usvc/go-config"

var conf = config.Map{
	"allowed-hostnames": &config.StringSlice{
		Default:   []string{"localhost"},
		Usage:     "hostnames that are allowed to connect to the websocket",
		Shorthand: "H",
	},
	"terminal-command": &config.String{
		Default:   "/bin/bash",
		Usage:     "absolute path to terminal",
		Shorthand: "t",
	},
	"terminal-args": &config.StringSlice{
		Default:   []string{"-l"},
		Usage:     "arguments to pass to terminal application",
		Shorthand: "r",
	},
	"max-buffer-size-bytes": &config.Int{
		Default:   512,
		Usage:     "maximum length of input from terminal",
		Shorthand: "B",
	},
	"workdir": &config.String{
		Default:   ".",
		Usage:     "working directory",
		Shorthand: "w",
	},
	"server-addr": &config.String{
		Default:   "0.0.0.0",
		Usage:     "ip interface the server should listen on",
		Shorthand: "a",
	},
	"server-port": &config.Int{
		Default:   8376,
		Usage:     "port the server should listen on",
		Shorthand: "p",
	},
}
