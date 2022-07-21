package cmd

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewConnInitCmd_noArgs(t *testing.T) {
	cmd := newConnInitCmd()
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err == nil {
		t.Error("expected command to fail")
	}
}

func TestNewConnInitCmd_emptyName(t *testing.T) {
	cmd := newConnInitCmd()

	b := new(bytes.Buffer)
	cmd.SetErr(b)
	cmd.SetArgs([]string{""})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("error execute cmd: %v", err)
	}

	out, err := io.ReadAll(b)
	if err != nil {
		t.Fatalf("error read out: %v", err)
	}

	if !strings.Contains(string(out), "connector name cannot be empty") {
		t.Errorf("expected: %s, actual: %s", "Error: connector name cannot be empty", string(out))
	}
}

func TestNewConnInitCmd(t *testing.T) {
	cmd := newConnInitCmd()

	testProjName := "test-connector-project"

	b := new(bytes.Buffer)
	cmd.SetOut(b)
	cmd.SetArgs([]string{testProjName})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("error execute cmd: %v", err)
	}
	defer func(command *exec.Cmd) {
		_ = command.Run()
	}(exec.Command("rm", "-rf", testProjName))

	expectedEntries := map[string]bool{
		testProjName: true,
		filepath.Join(testProjName, packageJsonName):            true,
		filepath.Join(testProjName, connectorSpecName):          true,
		filepath.Join(testProjName, "src"):                      true,
		filepath.Join(testProjName, "src", "index.spec.ts"):     true,
		filepath.Join(testProjName, "src", "index.ts"):          true,
		filepath.Join(testProjName, "src", "my-client.spec.ts"): true,
		filepath.Join(testProjName, "src", "my-client.ts"):      true,
		filepath.Join(testProjName, "tsconfig.json"):            true,
		filepath.Join(testProjName, ".gitignore"):               true,
	}

	numEntries := 0
	err := filepath.Walk(filepath.Join(".", testProjName),
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if _, ok := expectedEntries[path]; !ok {
				t.Errorf("error file not created: %s", path)
			}
			numEntries++

			return nil
		})
	if err != nil {
		t.Errorf("error walk '%s' dir: %v", testProjName, err)
	}

	if numEntries != len(expectedEntries) {
		t.Errorf("expected entries: %d, actual: %d", len(expectedEntries), numEntries)
	}

	out, err := io.ReadAll(b)
	if err != nil {
		t.Fatalf("error read out: %v", err)
	}

	if !strings.Contains(string(out), "Successfully created project") {
		t.Errorf("expected out to contain '%s', actual: %s", "Successfully created project", string(out))
	}
}
