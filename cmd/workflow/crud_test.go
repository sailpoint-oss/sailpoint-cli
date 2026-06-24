package workflow

import (
	"bytes"
	"context"
	"testing"

	beta "github.com/sailpoint-oss/golang-sdk/v2/api_beta"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/testutil"
)

func TestWorkflowCRUD(t *testing.T) {
	testutil.RequireLiveCredentials(t)
	testutil.SetJSONOutput(t)

	owner := requireWorkflowOwner(t)
	name := testutil.UniqueName("workflow")
	updatedDescription := "updated by sail CLI live CRUD test"
	dir := t.TempDir()

	createPath := testutil.WriteJSON(t, dir, "workflow-create.json", map[string]any{
		"name":        name,
		"description": "created by sail CLI live CRUD test",
		"enabled":     false,
		"owner": map[string]any{
			"type": "IDENTITY",
			"id":   owner.Id,
			"name": owner.Name,
		},
	})

	createCmd := newCreateCommand()
	createOut := new(bytes.Buffer)
	createCmd.SetOut(createOut)
	createCmd.SetArgs([]string{createPath})
	createCmd.Flags().Set("file", "true")

	if err := createCmd.Execute(); err != nil {
		testutil.SkipIfFeatureUnavailable(t, err)
		t.Fatalf("workflow create failed: %v", err)
	}

	created := testutil.DecodeJSON[[]beta.Workflow](t, createOut.String())
	if len(created) != 1 || created[0].GetId() == "" {
		t.Fatalf("expected one created workflow with an ID, got %#v", created)
	}
	workflowID := created[0].GetId()
	defer deleteWorkflow(t, workflowID)

	getWorkflowAndAssert(t, workflowID, name)

	updatePath := testutil.WriteJSON(t, dir, "workflow-update.json", created[0])
	updated := created[0]
	updated.SetDescription(updatedDescription)
	updatePath = testutil.WriteJSON(t, dir, "workflow-update.json", updated)

	updateCmd := newUpdateCommand()
	updateOut := new(bytes.Buffer)
	updateCmd.SetOut(updateOut)
	updateCmd.SetArgs([]string{updatePath})
	updateCmd.Flags().Set("file", "true")

	if err := updateCmd.Execute(); err != nil {
		t.Fatalf("workflow update failed: %v", err)
	}

	updatedWorkflow := testutil.DecodeJSON[beta.Workflow](t, updateOut.String())
	if updatedWorkflow.GetDescription() != updatedDescription {
		t.Fatalf("expected updated description %q, got %q", updatedDescription, updatedWorkflow.GetDescription())
	}

	getWorkflowAndAssert(t, workflowID, name)
}

type workflowOwner struct {
	Id   string
	Name string
}

func requireWorkflowOwner(t *testing.T) workflowOwner {
	t.Helper()

	apiClient, err := config.InitAPIClient(false)
	if err != nil {
		t.Fatalf("failed to initialize API client: %v", err)
	}
	identities, resp, err := apiClient.V2024.IdentitiesAPI.ListIdentities(context.TODO()).Limit(1).Execute()
	if err != nil {
		testutil.SkipIfFeatureUnavailable(t, err)
		t.Fatalf("failed to list identities for workflow owner: %v (response: %v)", err, resp)
	}
	if len(identities) == 0 || identities[0].GetId() == "" {
		t.Skip("skipping workflow CRUD test: no identity available to use as workflow owner")
	}
	return workflowOwner{Id: identities[0].GetId(), Name: identities[0].GetName()}
}

func getWorkflowAndAssert(t *testing.T, workflowID string, expectedName string) {
	t.Helper()

	getCmd := newGetCommand()
	getOut := new(bytes.Buffer)
	getCmd.SetOut(getOut)
	getCmd.SetArgs([]string{workflowID})
	if err := getCmd.Execute(); err != nil {
		t.Fatalf("workflow get failed: %v", err)
	}
	workflows := testutil.DecodeJSON[[]beta.Workflow](t, getOut.String())
	if len(workflows) != 1 {
		t.Fatalf("expected one workflow from get, got %d", len(workflows))
	}
	if workflows[0].GetId() != workflowID {
		t.Fatalf("expected workflow ID %q, got %q", workflowID, workflows[0].GetId())
	}
	if workflows[0].GetName() != expectedName {
		t.Fatalf("expected workflow name %q, got %q", expectedName, workflows[0].GetName())
	}
}

func deleteWorkflow(t *testing.T, workflowID string) {
	t.Helper()

	deleteCmd := newDeleteCommand()
	deleteCmd.SetArgs([]string{workflowID})
	if err := deleteCmd.Execute(); err != nil {
		t.Logf("failed to clean up workflow %s: %v", workflowID, err)
	}
}
