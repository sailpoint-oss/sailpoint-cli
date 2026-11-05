package api

// The automation surface deliberately does not use config.GetAuthToken: a
// cached PAT/OAuth token belongs to a CLI environment, not necessarily to the
// complete credential triple supplied by an unattended caller.
import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type requestResponse struct {
	Format  string      `json:"format"`
	Status  int         `json:"status"`
	Headers http.Header `json:"headers"`
	Body    []byte      `json:"body"` // JSON base64, preserving arbitrary bytes
}

func newRequestCmd() *cobra.Command {
	var bodyFile, contentType string
	var headers []string
	var timeout time.Duration
	cmd := &cobra.Command{
		Use:   "request METHOD /endpoint",
		Short: "Make one stateless request with a machine-readable HTTP response",
		Long: `Use exactly SAIL_BASE_URL, SAIL_CLIENT_ID and SAIL_CLIENT_SECRET.
Does not read or write CLI configuration, consult a keyring/token cache, retry,
follow redirects, or prompt. The timeout covers authentication and the request.
Outputs sail-api-response/1 JSON: status, headers and a base64 body. An HTTP
refusal is a response, not a process failure; callers must check status.
The request body file is streamed. Credentials and tokens never enter argv.`,
		Args:        cobra.ExactArgs(2),
		Annotations: map[string]string{"sailpoint.stateless": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().Changed("env") || cmd.Flags().Changed("debug") {
				return fmt.Errorf("request does not accept --env or --debug")
			}
			if timeout <= 0 || timeout > time.Hour {
				return fmt.Errorf("timeout must be positive and at most one hour")
			}
			base, err := url.Parse(os.Getenv("SAIL_BASE_URL"))
			if err != nil || base.Host == "" || base.User != nil || base.RawQuery != "" || base.Fragment != "" || (base.Path != "" && base.Path != "/") || (base.Scheme != "https" && !(base.Scheme == "http" && isLocalRequestHost(base.Hostname()))) {
				return fmt.Errorf("SAIL_BASE_URL must be an HTTPS origin (HTTP loopback is allowed for testing)")
			}
			id, secret := os.Getenv("SAIL_CLIENT_ID"), os.Getenv("SAIL_CLIENT_SECRET")
			if id == "" || secret == "" {
				return fmt.Errorf("request requires the complete SAIL_* credential triple")
			}
			method := strings.ToUpper(args[0])
			switch method {
			case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS":
			default:
				return fmt.Errorf("unsupported HTTP method")
			}
			endpoint, err := url.Parse(args[1])
			if err != nil || !strings.HasPrefix(args[1], "/") || strings.HasPrefix(args[1], "//") || endpoint.IsAbs() || endpoint.Host != "" || endpoint.Fragment != "" {
				return fmt.Errorf("endpoint must be a relative /path on SAIL_BASE_URL")
			}
			var body io.Reader
			var bodyLength int64
			if bodyFile != "" {
				f, err := os.Open(bodyFile)
				if err != nil {
					return fmt.Errorf("cannot open request body file")
				}
				defer f.Close()
				info, err := f.Stat()
				if err != nil || !info.Mode().IsRegular() {
					return fmt.Errorf("body file must be regular")
				}
				body = f
				bodyLength = info.Size()
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
			defer cancel()
			client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
			origin := strings.TrimSuffix(base.String(), "/")
			form := url.Values{"grant_type": {"client_credentials"}, "client_id": {id}, "client_secret": {secret}}
			auth, err := http.NewRequestWithContext(ctx, "POST", origin+"/oauth/token", strings.NewReader(form.Encode()))
			if err != nil {
				return fmt.Errorf("cannot construct authentication request")
			}
			auth.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			res, err := client.Do(auth)
			if err != nil {
				return fmt.Errorf("authentication transport failed")
			}
			var token struct {
				AccessToken string `json:"access_token"`
			}
			err = json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&token)
			res.Body.Close()
			if res.StatusCode != 200 || err != nil || token.AccessToken == "" || strings.ContainsAny(token.AccessToken, "\r\n") {
				return fmt.Errorf("authentication failed (HTTP %d)", res.StatusCode)
			}
			req, err := http.NewRequestWithContext(ctx, method, origin+args[1], body)
			if err != nil {
				return fmt.Errorf("cannot construct API request")
			}
			req.Header.Set("Authorization", "Bearer "+token.AccessToken)
			req.ContentLength = bodyLength
			req.Header.Set("Accept", "application/json")
			if body != nil {
				req.Header.Set("Content-Type", contentType)
			}
			for _, header := range headers {
				name, value, found := strings.Cut(header, ":")
				name, value = strings.TrimSpace(name), strings.TrimSpace(value)
				if !found || name == "" || strings.ContainsAny(name+value, "\r\n") {
					return fmt.Errorf("invalid request header")
				}
				switch strings.ToLower(name) {
				case "authorization", "proxy-authorization", "cookie", "host", "content-length", "transfer-encoding":
					return fmt.Errorf("reserved request header")
				}
				req.Header.Set(name, value)
			}
			res, err = client.Do(req)
			if err != nil {
				return fmt.Errorf("API transport failed; request outcome may be unknown")
			}
			defer res.Body.Close()
			raw, err := io.ReadAll(res.Body)
			if err != nil {
				return fmt.Errorf("API response incomplete; request outcome may be unknown")
			}
			return json.NewEncoder(cmd.OutOrStdout()).Encode(requestResponse{"sail-api-response/1", res.StatusCode, res.Header, raw})
		},
	}
	cmd.Flags().StringVar(&bodyFile, "body-file", "", "Stream request bytes from a regular file")
	cmd.Flags().StringArrayVar(&headers, "header", nil, "Additional non-authentication header (Key: Value)")
	cmd.Flags().StringVar(&contentType, "content-type", "application/json", "Request content type")
	cmd.Flags().DurationVar(&timeout, "timeout", 120*time.Second, "Whole operation deadline, including authentication")
	return cmd
}

func isLocalRequestHost(host string) bool {
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}
