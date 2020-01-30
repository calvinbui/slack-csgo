package rcon

import (
	"fmt"
	"github.com/james4k/rcon"
	"io"
	"log"
	"os"
)

// RCON is a remote console
type RCON struct {
	*rcon.RemoteConsole
}

// New makes a connection to the RCON server
func New(host, password string) (RCON, error) {
	remoteConsole, err := rcon.Dial(host, password)
	if err != nil {
		log.Fatal(err)
	}
	defer remoteConsole.Close()

	r := &RCON{remoteConsole}

	return *r, nil
}

// Execute runs rcon commands
func (r RCON) Execute(command string) (resp string) {
	reqID, err := r.RemoteConsole.Write(command)
	resp, respReqID, err := r.RemoteConsole.Read()

	if err != nil {
		if err == io.EOF {
			return
		}
		fmt.Fprintln(os.Stderr, "Failed to read command:", err.Error())
	}

	if reqID != respReqID {
		fmt.Println("Weird, This response is for another request")
	}

	return resp
}

// ChangeMap changes the map
func (r RCON) ChangeMap(m string) {
	r.Execute(m)
}
