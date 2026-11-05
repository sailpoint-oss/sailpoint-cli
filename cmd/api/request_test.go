package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func executeRequest(t *testing.T, args ...string) (requestResponse, string, error) {
	t.Helper()
	cmd := newRequestCmd()
	out, logs := new(bytes.Buffer), new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetErr(logs)
	cmd.SetArgs(args)
	err := cmd.Execute()
	var result requestResponse
	if err == nil {
		if decodeErr := json.Unmarshal(out.Bytes(), &result); decodeErr != nil {
			t.Fatal(decodeErr)
		}
	}
	return result, logs.String(), err
}

func TestStatelessRequestMultipartAndIdentity(t *testing.T) {
	payload := []byte("id,name\r\n1,Renée\r\n")
	file := filepath.Join(t.TempDir(), "a file.csv")
	if err := os.WriteFile(file, payload, 0600); err != nil {
		t.Fatal(err)
	}
	auths, requests := 0, 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth/token" {
			auths++
			if err := r.ParseForm(); err != nil {
				t.Error(err)
			}
			if r.Form.Get("client_id") != "fixture" || r.Form.Get("client_secret") != "s&=+%ecret" || r.Form.Get("grant_type") != "client_credentials" {
				t.Errorf("credential triple changed")
			}
			io.WriteString(w, `{"access_token":"fixture-token"}`)
			return
		}
		requests++
		if r.Header.Get("Authorization") != "Bearer fixture-token" || r.Header.Get("Content-Type") != "multipart/form-data; boundary=fixture" || r.URL.RawQuery != "object-type=custom%20group" {
			t.Error("request binding changed")
		}
		raw, _ := io.ReadAll(r.Body)
		if !bytes.Equal(raw, payload) {
			t.Error("file bytes changed")
		}
		w.Header().Set("Retry-After", "4")
		w.Header().Set("X-Total-Count", "37")
		w.WriteHeader(429)
		w.Write([]byte("refusal\r\n\x00"))
	}))
	defer server.Close()
	t.Setenv("SAIL_BASE_URL", server.URL)
	t.Setenv("SAIL_CLIENT_ID", "fixture")
	t.Setenv("SAIL_CLIENT_SECRET", "s&=+%ecret")
	t.Setenv("SAIL_DEBUG", "true")
	t.Setenv("SAIL_AUTHTYPE", "oauth")
	t.Setenv("SAIL_ACTIVEENVIRONMENT", "unrelated-cached-user")
	result, logs, err := executeRequest(t, "POST", "/beta/sources/one/load-entitlements?object-type=custom%20group", "--body-file", file, "--content-type", "multipart/form-data; boundary=fixture")
	if err != nil {
		t.Fatal(err)
	}
	if auths != 1 || requests != 1 || result.Format != "sail-api-response/1" || result.Status != 429 || result.Headers.Get("Retry-After") != "4" || result.Headers.Get("X-Total-Count") != "37" || string(result.Body) != "refusal\r\n\x00" {
		t.Fatalf("invalid response/counts: %+v %d %d", result, auths, requests)
	}
	if logs != "" {
		t.Fatalf("unexpected CLI log: %s", logs)
	}
}

func TestStatelessRequestRefusalsAndNoRedirectOrRetry(t *testing.T) {
	for _, status := range []int{200, 204, 400, 401, 403, 429, 500, 503, 307} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			calls, redirected := 0, 0
			target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { redirected++; w.WriteHeader(200) }))
			defer target.Close()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/oauth/token" {
					io.WriteString(w, `{"access_token":"test-token"}`)
					return
				}
				calls++
				w.Header().Set("Location", target.URL)
				w.WriteHeader(status)
				if status != 204 {
					io.WriteString(w, `{"result":"test"}`)
				}
			}))
			defer server.Close()
			t.Setenv("SAIL_BASE_URL", server.URL)
			t.Setenv("SAIL_CLIENT_ID", "test")
			t.Setenv("SAIL_CLIENT_SECRET", "test")
			result, _, err := executeRequest(t, "POST", "/test")
			if err != nil || result.Status != status || calls != 1 || redirected != 0 {
				t.Fatalf("result=%+v err=%v calls=%d redirected=%d", result, err, calls, redirected)
			}
		})
	}
}

func TestStatelessRequestValidationAndDeadline(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		time.Sleep(50 * time.Millisecond)
	}))
	defer server.Close()
	t.Setenv("SAIL_BASE_URL", server.URL)
	t.Setenv("SAIL_CLIENT_ID", "test")
	t.Setenv("SAIL_CLIENT_SECRET", "never-log-me")
	for _, args := range [][]string{{"GET", "https://other.example/test"}, {"GET", "//other.example/test"}, {"GET", "/test#fragment"}, {"UNKNOWN", "/test"}, {"GET", "/test", "--timeout", "0s"}} {
		_, logs, err := executeRequest(t, args...)
		if err == nil || calls != 0 || strings.Contains(logs, "never-log-me") {
			t.Fatalf("validation failed: %v %v", args, err)
		}
	}
	started := time.Now()
	_, logs, err := executeRequest(t, "POST", "/test", "--timeout", "20ms")
	if err == nil || time.Since(started) > time.Second || strings.Contains(logs, "never-log-me") {
		t.Fatalf("deadline/secret handling failed: %v", err)
	}
}
