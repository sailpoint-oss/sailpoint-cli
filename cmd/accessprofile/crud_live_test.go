package accessprofile

import (
	"bytes"
	"testing"

	v2024 "github.com/sailpoint-oss/golang-sdk/v2/api_v2024"
	"github.com/sailpoint-oss/sailpoint-cli/internal/testutil"
)

func TestAccessProfileCRUD(t *testing.T) {
	testutil.RequireLiveCredentials(t)
	testutil.SetJSONOutput(t)

	owner := testutil.FirstIdentity(t)
	source := testutil.FirstSource(t)
	name := testutil.UniqueName("access-profile")
	updatedDescription := "updated by sail CLI live CRUD test"
	dir := t.TempDir()

	createPath := testutil.WriteJSON(t, dir, "access-profile-create.json", map[string]any{
		"name":        name,
		"description": "created by sail CLI live CRUD test",
		"enabled":     false,
		"owner": map[string]any{
			"type": "IDENTITY",
			"id":   owner.ID,
		},
		"source": map[string]any{
			"type": "SOURCE",
			"id":   source.ID,
			"name": source.Name,
		},
		"entitlements": []any{},
	})

	createCmd := newCreateCommand()
	createOut := new(bytes.Buffer)
	createCmd.SetOut(createOut)
	createCmd.Flags().Set("file", createPath)

	if err := createCmd.Execute(); err != nil {
		testutil.SkipIfFeatureUnavailable(t, err)
		t.Fatalf("access-profile create failed: %v", err)
	}

	created := testutil.DecodeJSON[v2024.AccessProfile](t, createOut.String())
	if created.GetId() == "" {
		t.Fatalf("expected created access profile ID, got %#v", created)
	}
	accessProfileID := created.GetId()
	defer deleteAccessProfile(t, accessProfileID)

	getAccessProfileAndAssert(t, accessProfileID, name)
	listAccessProfileAndAssert(t, accessProfileID, name)

	patchPath := testutil.WriteJSON(t, dir, "access-profile-patch.json", testutil.StringPatch("/description", updatedDescription))
	patchCmd := newPatchCommand()
	patchOut := new(bytes.Buffer)
	patchCmd.SetOut(patchOut)
	patchCmd.SetArgs([]string{accessProfileID})
	patchCmd.Flags().Set("file", patchPath)

	if err := patchCmd.Execute(); err != nil {
		t.Fatalf("access-profile patch failed: %v", err)
	}
	updated := testutil.DecodeJSON[v2024.AccessProfile](t, patchOut.String())
	if updated.GetDescription() != updatedDescription {
		t.Fatalf("expected access profile description %q, got %q", updatedDescription, updated.GetDescription())
	}
}

func getAccessProfileAndAssert(t *testing.T, accessProfileID string, expectedName string) {
	t.Helper()

	getCmd := newGetCommand()
	getOut := new(bytes.Buffer)
	getCmd.SetOut(getOut)
	getCmd.SetArgs([]string{accessProfileID})
	if err := getCmd.Execute(); err != nil {
		t.Fatalf("access-profile get failed: %v", err)
	}
	profile := testutil.DecodeJSON[v2024.AccessProfile](t, getOut.String())
	if profile.GetId() != accessProfileID {
		t.Fatalf("expected access profile ID %q, got %q", accessProfileID, profile.GetId())
	}
	if profile.GetName() != expectedName {
		t.Fatalf("expected access profile name %q, got %q", expectedName, profile.GetName())
	}
}

func listAccessProfileAndAssert(t *testing.T, accessProfileID string, expectedName string) {
	t.Helper()

	listCmd := newListCommand()
	listOut := new(bytes.Buffer)
	listCmd.SetOut(listOut)
	listCmd.Flags().Set("filter", `id eq "`+accessProfileID+`"`)
	if err := listCmd.Execute(); err != nil {
		t.Fatalf("access-profile list failed: %v", err)
	}
	profiles := testutil.DecodeJSON[[]v2024.AccessProfile](t, listOut.String())
	if len(profiles) != 1 {
		t.Fatalf("expected one access profile from filtered list, got %d", len(profiles))
	}
	if profiles[0].GetName() != expectedName {
		t.Fatalf("expected listed access profile name %q, got %q", expectedName, profiles[0].GetName())
	}
}

func deleteAccessProfile(t *testing.T, accessProfileID string) {
	t.Helper()

	deleteCmd := newDeleteCommand()
	deleteCmd.SetArgs([]string{accessProfileID})
	deleteCmd.Flags().Set("force", "true")
	if err := deleteCmd.Execute(); err != nil {
		t.Logf("failed to clean up access profile %s: %v", accessProfileID, err)
	}
}
