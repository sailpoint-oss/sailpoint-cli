// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package va

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/pkg/sftp"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
	"golang.org/x/crypto/ssh"
)

func RunVACmd(addr string, password string, cmd string) (string, error) {
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

func RunVACmdLive(addr string, password string, cmd string) error {
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
		return dialErr
	}

	// Create a session. It is one session per command.
	session, sessionErr := client.NewSession()
	if sessionErr != nil {
		return sessionErr
	}
	defer session.Close()

	// get output
	session.Stdout = os.Stdout

	// Finally, run the command
	runErr := session.Run(cmd)
	if runErr != nil {
		return runErr
	}

	// Return the output
	return nil
}

func CollectVAFiles(endpoint string, password string, output string, files []string, p *mpb.Progress) error {
	log.Info("Starting File Collection", "VA", endpoint)

	config := &ssh.ClientConfig{
		User:            "sailpoint",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
	}

	outputFolder := path.Join(output, endpoint)

	// Connect
	client, err := ssh.Dial("tcp", net.JoinHostPort(endpoint, "22"), config)
	if err != nil {
		return fmt.Errorf("failed to dial: %v", err)
	}
	defer client.Close()

	sftp, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %v", err)
	}
	defer sftp.Close()

	var wg sync.WaitGroup
	for _, filePath := range files {
		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()

			err := collectFile(sftp, filePath, outputFolder, endpoint, p)
			if err != nil {
				log.Warn("Skipping file", "file", filePath, "VA", endpoint)
				log.Debug("Error collecting file", "file", filePath, "VA", endpoint, "err", err)
			}
		}(filePath)
	}

	wg.Wait()

	return nil
}

func collectFile(sftp *sftp.Client, filePath, outputFolder, endpoint string, p *mpb.Progress) error {
	_, base := path.Split(filePath)

	outputFile := path.Join(outputFolder, base)

	if _, err := os.Stat(outputFolder); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(outputFolder, 0700)
		if err != nil {
			return fmt.Errorf("failed to create output folder: %v", err)
		}
	}
	remoteFileStats, statErr := sftp.Stat(filePath)
	if statErr != nil {
		return fmt.Errorf("failed to stat remote file: %v", statErr)
	}

	name := fmt.Sprintf("%v - %v", endpoint, base)
	bar := p.AddBar(remoteFileStats.Size(),
		mpb.BarFillerClearOnComplete(),
		mpb.PrependDecorators(
			decor.Name(name, decor.WCSyncWidthR),
			decor.Name(" : ", decor.WCSyncWidthR),
			// decor.OnComplete(decor.CountersKiloByte("% .2f / % .2f", decor.WCSyncSpaceR), "Complete"),
			decor.TotalKiloByte("% .2f", decor.WCSyncSpaceR),
		),
		mpb.AppendDecorators(
			decor.OnComplete(decor.CountersKibiByte("% .2f / % .2f", decor.WCSyncWidth), "Complete")),
	)

	srcFile, err := sftp.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open remote file: %v", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create local file: %v", err)
	}
	defer dstFile.Close()

	writer := io.Writer(dstFile)
	proxyWriter := bar.ProxyWriter(writer)
	defer proxyWriter.Close()

	_, err = io.Copy(proxyWriter, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file contents: %v", err)
	}

	return nil
}
