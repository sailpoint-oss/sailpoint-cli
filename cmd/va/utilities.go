package va

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"strings"
	"syscall"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

func password() (string, error) {
	fmt.Print("Enter Password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(bytePassword)), nil
}

func runVACmd(addr string, password string, cmd string) (string, error) {
	config := &ssh.ClientConfig{
		User:            "sailpoint",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
	}
	// Connect
	client, err := ssh.Dial("tcp", net.JoinHostPort(addr, "22"), config)
	if err != nil {
		return "", err
	}

	// Create a session. It is one session per command.
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	// import "bytes"
	var b bytes.Buffer

	// get output
	session.Stdout = &b

	// Finally, run the command
	err = session.Run(cmd)

	// Return the output
	return b.String(), err
}

func getVAFile(addr string, password string, remoteFile string, outputDir string) error {
	config := &ssh.ClientConfig{
		User:            "sailpoint",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
	}

	_, base := path.Split(remoteFile)
	outputDir = outputDir + "/" + base
	// Connect
	client, err := ssh.Dial("tcp", net.JoinHostPort(addr, "22"), config)
	if err != nil {
		return err
	}

	// Create a session. It is one session per command.
	sftp, err := sftp.NewClient(client)
	if err != nil {
		return err
	}
	defer sftp.Close()

	// Open the source file
	srcFile, err := sftp.Open(remoteFile)
	if err != nil {
		log.Fatal(err)
	}
	defer srcFile.Close()

	// Create the destination file
	dstFile, err := os.Create(outputDir)
	if err != nil {
		log.Fatal(err)
	}
	defer dstFile.Close()

	// Copy the file
	bytesWritten, writeErr := srcFile.WriteTo(dstFile)
	if writeErr != nil {
		return writeErr
	}
	fmt.Printf("\nSaved %v to %v (%v bytes)", base, outputDir, bytesWritten)

	return nil
}
