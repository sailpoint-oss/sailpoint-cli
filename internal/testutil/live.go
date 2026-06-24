package testutil

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	v2024 "github.com/sailpoint-oss/golang-sdk/v2/api_v2024"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/spf13/viper"
)

func RequireLiveCredentials(t *testing.T) {
	t.Helper()

	if err := config.InitConfig(); err != nil {
		t.Fatalf("failed to initialize CLI config: %v", err)
	}
	if err := config.Validate(); err != nil {
		t.Skipf("skipping live CLI test: no usable SailPoint CLI credentials found (%v). Configure PAT credentials with SAIL_BASE_URL, SAIL_CLIENT_ID, and SAIL_CLIENT_SECRET, or run `sail env create`/`sail auth login` for OAuth, then rerun this test.", err)
	}
}

func SetJSONOutput(t *testing.T) {
	t.Helper()

	previousJSON := viper.GetBool("json")
	previousOutput := viper.GetString("output")
	viper.Set("json", false)
	viper.Set("output", "json")
	t.Cleanup(func() {
		viper.Set("json", previousJSON)
		viper.Set("output", previousOutput)
	})
}

func UniqueName(prefix string) string {
	sanitized := strings.Trim(strings.ToLower(prefix), "-")
	if sanitized == "" {
		sanitized = "resource"
	}
	return fmt.Sprintf("sail-cli-ci-%s-%d", sanitized, time.Now().UnixNano())
}

func WriteJSON(t *testing.T, dir string, name string, value any) string {
	t.Helper()

	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal %s: %v", name, err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("failed to write %s: %v", name, err)
	}
	return path
}

func DecodeJSON[T any](t *testing.T, raw string) T {
	t.Helper()

	var value T
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		t.Fatalf("failed to decode JSON output %q: %v", raw, err)
	}
	return value
}

func SkipIfFeatureUnavailable(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		return
	}
	message := err.Error()
	for _, status := range []string{"401", "403", "404", "not found", "forbidden", "unauthorized"} {
		if strings.Contains(strings.ToLower(message), status) {
			t.Skipf("skipping live CLI test: required tenant feature or permission unavailable: %v", err)
		}
	}
}

type IdentityRef struct {
	ID   string
	Name string
}

type SourceRef struct {
	ID   string
	Name string
}

func FirstIdentity(t *testing.T) IdentityRef {
	t.Helper()

	apiClient, err := config.InitAPIClient(false)
	if err != nil {
		t.Fatalf("failed to initialize API client: %v", err)
	}
	identities, resp, err := apiClient.V2024.IdentitiesAPI.ListIdentities(context.TODO()).Limit(1).Execute()
	if err != nil {
		SkipIfFeatureUnavailable(t, err)
		t.Fatalf("failed to list identities for live test owner: %v (response: %v)", err, resp)
	}
	if len(identities) == 0 || identities[0].GetId() == "" {
		t.Skip("skipping live CLI test: no identity available to use as owner")
	}
	return IdentityRef{ID: identities[0].GetId(), Name: identities[0].GetName()}
}

func FirstSource(t *testing.T) SourceRef {
	t.Helper()

	apiClient, err := config.InitAPIClient(false)
	if err != nil {
		t.Fatalf("failed to initialize API client: %v", err)
	}
	sources, resp, err := apiClient.V2024.SourcesAPI.ListSources(context.TODO()).Limit(1).Execute()
	if err != nil {
		SkipIfFeatureUnavailable(t, err)
		t.Fatalf("failed to list sources for live test fixture: %v (response: %v)", err, resp)
	}
	if len(sources) == 0 || sources[0].GetId() == "" {
		t.Skip("skipping live CLI test: no source available to use as fixture")
	}
	return SourceRef{ID: sources[0].GetId(), Name: sources[0].GetName()}
}

func StringPatch(path string, value string) []v2024.JsonPatchOperation {
	return []v2024.JsonPatchOperation{
		{
			Op:    "replace",
			Path:  path,
			Value: &v2024.UpdateMultiHostSourcesRequestInnerValue{String: &value},
		},
	}
}
