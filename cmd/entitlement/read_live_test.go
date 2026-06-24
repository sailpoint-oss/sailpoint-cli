package entitlement

import (
	"bytes"
	"testing"

	v2024 "github.com/sailpoint-oss/golang-sdk/v2/api_v2024"
	"github.com/sailpoint-oss/sailpoint-cli/internal/testutil"
)

func TestEntitlementListAndGet(t *testing.T) {
	testutil.RequireLiveCredentials(t)
	testutil.SetJSONOutput(t)

	listCmd := newListCommand()
	listOut := new(bytes.Buffer)
	listCmd.SetOut(listOut)
	listCmd.Flags().Set("limit", "5")

	if err := listCmd.Execute(); err != nil {
		testutil.SkipIfFeatureUnavailable(t, err)
		t.Fatalf("entitlement list failed: %v", err)
	}

	entitlements := testutil.DecodeJSON[[]v2024.Entitlement](t, listOut.String())
	if len(entitlements) == 0 {
		t.Skip("skipping entitlement get test: entitlement list returned no results")
	}

	entitlementID := entitlements[0].GetId()
	if entitlementID == "" {
		t.Fatalf("expected listed entitlement to have an ID: %#v", entitlements[0])
	}

	getCmd := newGetCommand()
	getOut := new(bytes.Buffer)
	getCmd.SetOut(getOut)
	getCmd.SetArgs([]string{entitlementID})

	if err := getCmd.Execute(); err != nil {
		t.Fatalf("entitlement get failed: %v", err)
	}

	entitlement := testutil.DecodeJSON[v2024.Entitlement](t, getOut.String())
	if entitlement.GetId() != entitlementID {
		t.Fatalf("expected entitlement ID %q, got %q", entitlementID, entitlement.GetId())
	}
}
