// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package va

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/pkg/sftp"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

const sshDirMode = 0700

// ensureKnownHostsFile creates ~/.ssh and an empty known_hosts file if missing.
func ensureKnownHostsFile(knownHostsPath string) error {
	sshDir := path.Dir(knownHostsPath)
	if err := os.MkdirAll(sshDir, sshDirMode); err != nil {
		return fmt.Errorf("could not create %s: %w", sshDir, err)
	}
	f, err := os.OpenFile(knownHostsPath, os.O_CREATE|os.O_RDONLY, 0600)
	if err != nil {
		return fmt.Errorf("could not create %s: %w", knownHostsPath, err)
	}
	_ = f.Close()
	return nil
}

// promptYesNo reads a line from stdin and returns true for yes, false for no or invalid input.
func promptYesNo() bool {
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return false
	}
	line := strings.TrimSpace(strings.ToLower(scanner.Text()))
	return line == "yes" || line == "y"
}

// newSSHClientConfig returns an ssh.ClientConfig with host key verification via ~/.ssh/known_hosts
// and interactive first-time host acceptance.
func newSSHClientConfig(password string) (*ssh.ClientConfig, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not determine user home directory: %w", err)
	}
	knownHostsPath := path.Join(userHome, ".ssh", "known_hosts")
	if err := ensureKnownHostsFile(knownHostsPath); err != nil {
		return nil, err
	}

	baseCallback, err := knownhosts.New(knownHostsPath)
	if err != nil {
		return nil, fmt.Errorf("could not load SSH known_hosts from %s: %w", knownHostsPath, err)
	}

	callback := func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		err := baseCallback(hostname, remote, key)
		if err == nil {
			return nil
		}
		var keyErr *knownhosts.KeyError
		if errors.As(err, &keyErr) {
			// Host was known but key differs (possible MITM): always fail
			if len(keyErr.Want) > 0 {
				return err
			}
			// Unknown host (first-time connection): prompt and optionally add to known_hosts
			fingerprint := ssh.FingerprintSHA256(key)
			addrStr := hostname
			if remote != nil {
				addrStr = remote.String()
			}
			fmt.Fprintf(os.Stderr, "The authenticity of host %q can't be established.\n%s key fingerprint is %s.\nAre you sure you want to continue connecting (yes/no)? ", addrStr, key.Type(), fingerprint)
			if !promptYesNo() {
				return err
			}
			line := knownhosts.Line([]string{knownhosts.Normalize(addrStr)}, key) + "\n"
			f, appendErr := os.OpenFile(knownHostsPath, os.O_WRONLY|os.O_APPEND, 0600)
			if appendErr != nil {
				return fmt.Errorf("could not append to known_hosts: %w", appendErr)
			}
			_, _ = f.WriteString(line)
			_ = f.Close()
			return nil
		}
		return err
	}

	return &ssh.ClientConfig{
		User:            "sailpoint",
		HostKeyCallback: callback,
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
	}, nil
}

func RunVACmd(addr string, password string, cmd string) (string, error) {
	config, err := newSSHClientConfig(password)
	if err != nil {
		return "", err
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
	config, err := newSSHClientConfig(password)
	if err != nil {
		return err
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

	config, err := newSSHClientConfig(password)
	if err != nil {
		return err
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
