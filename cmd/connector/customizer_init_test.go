// connector/customizer_init_test.go
package connector

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewCustomizerInitCmd(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		preCreateDirs []string
		wantErr       bool
		wantOutSubstr string
		wantErrSubstr string
		verify        func(t *testing.T, workdir string)
	}{
		{
			name:          "success",
			args:          []string{"mycustomizer"},
			wantErr:       false,
			wantOutSubstr: "Successfully created project 'mycustomizer'",
			verify: func(t *testing.T, workdir string) {
				projectPath := filepath.Join(workdir, "mycustomizer")
				info, err := os.Stat(projectPath)
				if err != nil {
					t.Fatalf("expected project dir, got error: %v", err)
				}
				if !info.IsDir() {
					t.Errorf("expected %q to be a directory", projectPath)
				}
				if _, err := os.Stat(filepath.Join(projectPath, packageJsonName)); err != nil {
					t.Errorf("expected %s in %s, got error: %v", packageJsonName, projectPath, err)
				}
			},
		},
		{
			name:          "already exists",
			args:          []string{"exists"},
			preCreateDirs: []string{"exists"},
			wantErr:       false,
			wantErrSubstr: "project 'exists' already exists",
			verify: func(t *testing.T, workdir string) {
				if _, err := os.Stat(filepath.Join(workdir, "exists")); err != nil {
					t.Errorf("expected existing dir to remain, got error: %v", err)
				}
			},
		},
		{
			name:          "no args",
			args:          []string{},
			wantErr:       true,
			wantErrSubstr: "accepts 1 arg(s), received 0",
		},
	}

	for _, tc := range tests {
		tc := tc // capture
		t.Run(tc.name, func(t *testing.T) {
			// capture and restore cwd so temp dirs can be removed on Windows
			origWd, err := os.Getwd()
			if err != nil {
				t.Fatalf("could not get wd: %v", err)
			}
			defer func() {
				_ = os.Chdir(origWd)
			}()

			// switch into a temp dir for FS isolation
			workdir := origWd
			if tc.name != "no args" {
				workdir = t.TempDir()
				if err := os.Chdir(workdir); err != nil {
					t.Fatalf("chdir failed: %v", err)
				}
				// pre-create directories if needed
				for _, d := range tc.preCreateDirs {
					if err := os.Mkdir(d, 0o755); err != nil {
						t.Fatalf("pre-create dir %q: %v", d, err)
					}
				}
			}

			cmd := newCustomizerInitCmd()
			var outBuf, errBuf bytes.Buffer
			cmd.SetOut(&outBuf)
			cmd.SetErr(&errBuf)
			cmd.SetArgs(tc.args)

			err = cmd.Execute()
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil; stderr=%q", errBuf.String())
				}
				if !strings.Contains(err.Error(), tc.wantErrSubstr) &&
					!strings.Contains(errBuf.String(), tc.wantErrSubstr) {
					t.Errorf("error %q does not contain %q", err, tc.wantErrSubstr)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v; stderr=%q", err, errBuf.String())
				}
				out := outBuf.String()
				if tc.wantOutSubstr != "" && !strings.Contains(out, tc.wantOutSubstr) {
					t.Errorf("stdout %q does not contain %q", out, tc.wantOutSubstr)
				}
			}

			if tc.verify != nil {
				tc.verify(t, workdir)
			}
		})
	}
}
