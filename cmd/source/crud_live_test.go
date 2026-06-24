package source

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	v2024 "github.com/sailpoint-oss/golang-sdk/v2/api_v2024"
	"github.com/sailpoint-oss/sailpoint-cli/internal/testutil"
)

func TestSourceCRUD(t *testing.T) {
	testutil.RequireLiveCredentials(t)
	testutil.SetJSONOutput(t)

	fixturePath := os.Getenv("SAIL_TEST_SOURCE_CREATE_PAYLOAD")
	if fixturePath == "" {
		t.Skip("skipping source CRUD test: SAIL_TEST_SOURCE_CREATE_PAYLOAD is required for connector-specific source creation")
	}

	raw, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to read source fixture %q: %v", fixturePath, err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("failed to decode source fixture %q: %v", fixturePath, err)
	}

	dir := t.TempDir()
	name := testutil.UniqueName("source")
	updatedDescription := "updated by sail CLI live CRUD test"
	payload["name"] = name
	delete(payload, "id")
	createPath := testutil.WriteJSON(t, dir, "source-create.json", payload)

	createCmd := newCreateCommand()
	createOut := new(bytes.Buffer)
	createCmd.SetOut(createOut)
	createCmd.Flags().Set("file", createPath)

	if err := createCmd.Execute(); err != nil {
		testutil.SkipIfFeatureUnavailable(t, err)
		t.Fatalf("source create failed: %v", err)
	}

	created := testutil.DecodeJSON[v2024.Source](t, createOut.String())
	if created.GetId() == "" {
		t.Fatalf("expected created source ID, got %#v", created)
	}
	sourceID := created.GetId()
	defer deleteSource(t, sourceID)

	getSourceAndAssert(t, sourceID, name)

	patchPath := testutil.WriteJSON(t, dir, "source-patch.json", []map[string]any{
		{"op": "replace", "path": "/description", "value": updatedDescription},
	})
	patchCmd := newPatchCommand()
	patchOut := new(bytes.Buffer)
	patchCmd.SetOut(patchOut)
	patchCmd.SetArgs([]string{sourceID})
	patchCmd.Flags().Set("file", patchPath)

	if err := patchCmd.Execute(); err != nil {
		t.Fatalf("source patch failed: %v", err)
	}
	updated := testutil.DecodeJSON[v2024.Source](t, patchOut.String())
	if updated.GetDescription() != updatedDescription {
		t.Fatalf("expected source description %q, got %q", updatedDescription, updated.GetDescription())
	}
}

func getSourceAndAssert(t *testing.T, sourceID string, expectedName string) {
	t.Helper()

	getCmd := newGetCommand()
	getOut := new(bytes.Buffer)
	getCmd.SetOut(getOut)
	getCmd.SetArgs([]string{sourceID})
	if err := getCmd.Execute(); err != nil {
		t.Fatalf("source get failed: %v", err)
	}
	source := testutil.DecodeJSON[v2024.Source](t, getOut.String())
	if source.GetId() != sourceID {
		t.Fatalf("expected source ID %q, got %q", sourceID, source.GetId())
	}
	if source.GetName() != expectedName {
		t.Fatalf("expected source name %q, got %q", expectedName, source.GetName())
	}
}

func deleteSource(t *testing.T, sourceID string) {
	t.Helper()

	deleteCmd := newDeleteCommand()
	deleteCmd.SetArgs([]string{sourceID})
	deleteCmd.Flags().Set("force", "true")
	if err := deleteCmd.Execute(); err != nil {
		t.Logf("failed to clean up source %s: %v", sourceID, err)
	}
}
