package va

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"strings"
	"syscall"

	"github.com/fatih/color"
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
	fmt.Println()
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

func getVAFile(addr string, password string, remoteFile string, outputDir string) error {
	config := &ssh.ClientConfig{
		User:            "sailpoint",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
	}

	_, base := path.Split(remoteFile)
	outputDir = path.Join(outputDir, addr)
	outputFile := path.Join(outputDir, base)
	if _, err := os.Stat(outputDir); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(outputDir, 0700)
		if err != nil {
			return err
		}
	}

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
		return err
	}
	defer srcFile.Close()

	// Create the destination file
	dstFile, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy the file
	bytesWritten, writeErr := srcFile.WriteTo(dstFile)
	if writeErr != nil {
		return writeErr
	}

	color.Green("Saved %v to %v (%v bytes)", base, outputFile, bytesWritten)

	return nil
}
