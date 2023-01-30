// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package util

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path"
	"sync"
	"time"

	"github.com/fatih/color"
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

func CollectVAFiles(endpoint string, password string, output string, files []string) error {
	color.Blue("Starting File Collection for %s\n", endpoint)
	config := &ssh.ClientConfig{
		User:            "sailpoint",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
	}

	var wg sync.WaitGroup
	// passed wg will be accounted at p.Wait() call
	p := mpb.New(mpb.WithWidth(60),
		mpb.PopCompletedMode(),
		mpb.WithRefreshRate(180*time.Millisecond),
		mpb.WithWaitGroup(&wg))

	log.SetOutput(p)

	outputFolder := path.Join(output, endpoint)

	for i := 0; i < len(files); i++ {
		filePath := files[i]
		wg.Add(1)
		go func(filePath string) error {
			// Connect
			client, err := ssh.Dial("tcp", net.JoinHostPort(endpoint, "22"), config)
			if err != nil {
				fmt.Println(err)
			}

			sftp, err := sftp.NewClient(client)
			if err != nil {
				fmt.Println(err)
			}
			defer sftp.Close()

			defer wg.Done()

			_, base := path.Split(filePath)
			outputFile := path.Join(outputFolder, base)
			if _, err := os.Stat(outputFolder); errors.Is(err, os.ErrNotExist) {
				err := os.MkdirAll(outputFolder, 0700)
				if err != nil {
					fmt.Println(err)
				}
			}
			remoteFileStats, statErr := sftp.Stat(filePath)
			if statErr == nil {
				name := fmt.Sprintf("%v - %v", endpoint, base)
				bar := p.AddBar(remoteFileStats.Size(),
					mpb.BarFillerClearOnComplete(),
					mpb.PrependDecorators(
						// simple name decorator
						decor.Name(name, decor.WCSyncSpaceR),
						decor.Name(":", decor.WCSyncSpaceR),
						decor.OnComplete(decor.CountersKiloByte("% .2f / % .2f", decor.WCSyncSpaceR), "Complete"),
						decor.TotalKiloByte("% .2f", decor.WCSyncSpaceR),
					),
					mpb.AppendDecorators(
						decor.OnComplete(decor.EwmaSpeed(decor.UnitKB, "% .2f", 90, decor.WCSyncWidth), ""),
						decor.OnComplete(decor.Percentage(decor.WC{W: 5}), ""),
					),
				)

				// Open the source file
				srcFile, err := sftp.Open(filePath)
				if err != nil {
					log.Println(err)
				}
				defer srcFile.Close()

				// Create the destination file
				dstFile, err := os.Create(outputFile)
				if err != nil {
					return err
				}
				defer dstFile.Close()

				writer := io.Writer(dstFile)

				// create proxy reader
				proxyWriter := bar.ProxyWriter(writer)
				defer proxyWriter.Close()

				io.Copy(proxyWriter, srcFile)
			}
			return nil
		}(filePath)
	}

	p.Wait()

	return nil
}
