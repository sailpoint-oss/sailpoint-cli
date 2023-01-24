package va

import (
	"bytes"
	"net"

	"golang.org/x/crypto/ssh"
)

func runVACmd(addr string, password string, cmd string) (string, error) {
	config := &ssh.ClientConfig{
		User:            "sailpoint",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
	}
	// Connect
	client, dialErr := ssh.Dial("tcp", net.JoinHostPort(addr, "22"), config)
	if dialErr != nil {
		return "", dialErr
	}

	// Create a session. It is one session per command.
	session, sessionErr := client.NewSession()
	if sessionErr != nil {
		return "", sessionErr
	}
	defer session.Close()

	// import "bytes"
	var b bytes.Buffer

	// get output
	session.Stdout = &b

	// Finally, run the command
	runErr := session.Run(cmd)
	if runErr != nil {
		return b.String(), runErr
	}

	// Return the output
	return b.String(), nil
}
