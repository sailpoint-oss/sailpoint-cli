package role

import (
	"bytes"
	"testing"

	v2024 "github.com/sailpoint-oss/golang-sdk/v2/api_v2024"
	"github.com/sailpoint-oss/sailpoint-cli/internal/testutil"
)

func TestRoleCRUD(t *testing.T) {
	testutil.RequireLiveCredentials(t)
	testutil.SetJSONOutput(t)

	owner := testutil.FirstIdentity(t)
	name := testutil.UniqueName("role")
	updatedDescription := "updated by sail CLI live CRUD test"
	dir := t.TempDir()

	createPath := testutil.WriteJSON(t, dir, "role-create.json", map[string]any{
		"name":        name,
		"description": "created by sail CLI live CRUD test",
		"enabled":     false,
		"requestable": false,
		"owner": map[string]any{
			"type": "IDENTITY",
			"id":   owner.ID,
		},
	})

	createCmd := newCreateCommand()
	createOut := new(bytes.Buffer)
	createCmd.SetOut(createOut)
	createCmd.Flags().Set("file", createPath)

	if err := createCmd.Execute(); err != nil {
		testutil.SkipIfFeatureUnavailable(t, err)
		t.Fatalf("role create failed: %v", err)
	}

	created := testutil.DecodeJSON[v2024.Role](t, createOut.String())
	if created.GetId() == "" {
		t.Fatalf("expected created role ID, got %#v", created)
	}
	roleID := created.GetId()
	defer deleteRole(t, roleID)

	getRoleAndAssert(t, roleID, name)
	listRoleAndAssert(t, roleID, name)

	patchPath := testutil.WriteJSON(t, dir, "role-patch.json", testutil.StringPatch("/description", updatedDescription))
	patchCmd := newPatchCommand()
	patchOut := new(bytes.Buffer)
	patchCmd.SetOut(patchOut)
	patchCmd.SetArgs([]string{roleID})
	patchCmd.Flags().Set("file", patchPath)

	if err := patchCmd.Execute(); err != nil {
		t.Fatalf("role patch failed: %v", err)
	}
	updated := testutil.DecodeJSON[v2024.Role](t, patchOut.String())
	if updated.GetDescription() != updatedDescription {
		t.Fatalf("expected role description %q, got %q", updatedDescription, updated.GetDescription())
	}
}

func getRoleAndAssert(t *testing.T, roleID string, expectedName string) {
	t.Helper()

	getCmd := newGetCommand()
	getOut := new(bytes.Buffer)
	getCmd.SetOut(getOut)
	getCmd.SetArgs([]string{roleID})
	if err := getCmd.Execute(); err != nil {
		t.Fatalf("role get failed: %v", err)
	}
	role := testutil.DecodeJSON[v2024.Role](t, getOut.String())
	if role.GetId() != roleID {
		t.Fatalf("expected role ID %q, got %q", roleID, role.GetId())
	}
	if role.GetName() != expectedName {
		t.Fatalf("expected role name %q, got %q", expectedName, role.GetName())
	}
}

func listRoleAndAssert(t *testing.T, roleID string, expectedName string) {
	t.Helper()

	listCmd := newListCommand()
	listOut := new(bytes.Buffer)
	listCmd.SetOut(listOut)
	listCmd.Flags().Set("filter", `id eq "`+roleID+`"`)
	if err := listCmd.Execute(); err != nil {
		t.Fatalf("role list failed: %v", err)
	}
	roles := testutil.DecodeJSON[[]v2024.Role](t, listOut.String())
	if len(roles) != 1 {
		t.Fatalf("expected one role from filtered list, got %d", len(roles))
	}
	if roles[0].GetName() != expectedName {
		t.Fatalf("expected listed role name %q, got %q", expectedName, roles[0].GetName())
	}
}

func deleteRole(t *testing.T, roleID string) {
	t.Helper()

	deleteCmd := newDeleteCommand()
	deleteCmd.SetArgs([]string{roleID})
	deleteCmd.Flags().Set("force", "true")
	if err := deleteCmd.Execute(); err != nil {
		t.Logf("failed to clean up role %s: %v", roleID, err)
	}
}
