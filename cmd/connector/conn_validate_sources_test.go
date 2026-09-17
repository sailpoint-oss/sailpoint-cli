// Copyright (c) 2022, SailPoint Technologies, Inc. All rights reserved.
package connector

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseRepositoryOwner(t *testing.T) {
	for _, tt := range []struct {
		name       string
		repository string
		owner      string
		wantErr    bool
	}{
		{name: "https remote", repository: "https://github.com/sailpoint/saas-conn-freshservice", owner: "sailpoint"},
		{name: "https remote with suffix", repository: "https://github.com/sailpoint-oss/saas-conn-github.git", owner: "sailpoint-oss"},
		{name: "scp like remote", repository: "git@github.com:sailpoint/saas-conn-smartsheet.git", owner: "sailpoint"},
		{name: "ssh remote", repository: "ssh://git@github.com/sailpoint/saas-conn-lastpass.git", owner: "sailpoint"},
		{name: "surrounding space", repository: "  https://github.com/sailpoint/saas-conn-aha  ", owner: "sailpoint"},
		{name: "empty", repository: "", wantErr: true},
		{name: "option like value", repository: "--upload-pack=touch /tmp/pwned", wantErr: true},
		{name: "remote helper", repository: "ext::sh -c touch% /tmp/pwned", wantErr: true},
		{name: "local path", repository: "/tmp/evil-repo", wantErr: true},
		{name: "relative path", repository: "./evil-repo", wantErr: true},
		{name: "file transport", repository: "file:///tmp/evil-repo", wantErr: true},
		{name: "other host", repository: "https://gitlab.com/sailpoint/saas-conn-github.git", wantErr: true},
		{name: "host lookalike", repository: "https://github.com.attacker.test/sailpoint/saas-conn-github.git", wantErr: true},
		{name: "embedded credentials", repository: "https://attacker:token@github.com/sailpoint/saas-conn-github.git", wantErr: true},
		{name: "nested path", repository: "https://github.com/sailpoint/saas-conn-github/extra", wantErr: true},
		{name: "missing repo", repository: "https://github.com/sailpoint", wantErr: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			owner, err := parseRepositoryOwner(tt.repository)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error for %q, got owner %q", tt.repository, owner)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.repository, err)
			}
			if owner != tt.owner {
				t.Errorf("expected owner %q, got %q", tt.owner, owner)
			}
		})
	}
}

func TestCheckRepositoryRef(t *testing.T) {
	const sha = "3877ab4397617253e923f62f86423e1053e0fef9"

	for _, tt := range []struct {
		name            string
		ref             string
		allowMutableRef bool
		wantErr         bool
	}{
		{name: "commit sha", ref: sha},
		{name: "commit sha with mutable allowed", ref: sha, allowMutableRef: true},
		{name: "branch rejected by default", ref: "main", wantErr: true},
		{name: "short sha rejected by default", ref: sha[:12], wantErr: true},
		{name: "empty", ref: "", wantErr: true},
		{name: "empty with mutable allowed", ref: "", allowMutableRef: true, wantErr: true},
		{name: "branch allowed", ref: "main", allowMutableRef: true},
		{name: "namespaced branch allowed", ref: "release/2.5.0", allowMutableRef: true},
		{name: "option like ref", ref: "--upload-pack=id", allowMutableRef: true, wantErr: true},
		{name: "shell metacharacters", ref: "main; touch /tmp/pwned", allowMutableRef: true, wantErr: true},
		{name: "command substitution", ref: "$(touch /tmp/pwned)", allowMutableRef: true, wantErr: true},
		{name: "range syntax", ref: "main..evil", allowMutableRef: true, wantErr: true},
		{name: "reflog syntax", ref: "main@{1}", allowMutableRef: true, wantErr: true},
		{name: "lock suffix", ref: "main.lock", allowMutableRef: true, wantErr: true},
		{name: "trailing slash", ref: "release/", allowMutableRef: true, wantErr: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := checkRepositoryRef(tt.ref, tt.allowMutableRef)
			if tt.wantErr && err == nil {
				t.Fatalf("expected an error for ref %q", tt.ref)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error for ref %q: %v", tt.ref, err)
			}
		})
	}
}

func TestCheckSourcesTrusted(t *testing.T) {
	const sha = "3877ab4397617253e923f62f86423e1053e0fef9"

	trusted := Source{Name: "github", Repository: "git@github.com:sailpoint/saas-conn-github.git", RepositoryRef: sha}
	defaultPolicy := sourceTrustPolicy{allowedOwners: defaultAllowedRepositoryOwners}

	t.Run("trusted sources pass", func(t *testing.T) {
		if err := checkSourcesTrusted([]Source{trusted}, defaultPolicy); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("untrusted owner is rejected", func(t *testing.T) {
		untrusted := trusted
		untrusted.Repository = "https://github.com/attacker/saas-conn-github.git"

		err := checkSourcesTrusted([]Source{trusted, untrusted}, defaultPolicy)
		if err == nil {
			t.Fatal("expected an error for an untrusted owner")
		}
		if !strings.Contains(err.Error(), "attacker") {
			t.Errorf("expected the error to name the owner, got %v", err)
		}
	})

	t.Run("owner allowlist can be extended", func(t *testing.T) {
		extra := trusted
		extra.Repository = "https://github.com/partner-org/saas-conn-github.git"

		policy := sourceTrustPolicy{allowedOwners: append(defaultAllowedRepositoryOwners, "Partner-Org")}
		if err := checkSourcesTrusted([]Source{extra}, policy); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("unnamed entry is reported by position", func(t *testing.T) {
		unnamed := trusted
		unnamed.Name = ""

		err := checkSourcesTrusted([]Source{trusted, unnamed}, defaultPolicy)
		if err == nil {
			t.Fatal("expected an error for an unnamed entry")
		}
		if !strings.Contains(err.Error(), "entry 2") {
			t.Errorf("expected the error to name the entry position, got %v", err)
		}
	})
}

func TestConfirmSources(t *testing.T) {
	sources := []Source{{Name: "github", Repository: "git@github.com:sailpoint/saas-conn-github.git", RepositoryRef: "main"}}

	for _, tt := range []struct {
		name  string
		input string
		want  bool
	}{
		{name: "y", input: "y\n", want: true},
		{name: "yes uppercase", input: "YES\n", want: true},
		{name: "n", input: "n\n", want: false},
		{name: "empty defaults to no", input: "\n", want: false},
		{name: "no input defaults to no", input: "", want: false},
		{name: "unrelated answer defaults to no", input: "sure\n", want: false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer

			confirmed, err := confirmSources(strings.NewReader(tt.input), &out, sources)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if confirmed != tt.want {
				t.Errorf("expected %v, got %v", tt.want, confirmed)
			}
			if !strings.Contains(out.String(), "saas-conn-github") {
				t.Errorf("expected the prompt to list the repository, got %q", out.String())
			}
		})
	}
}
