package connector

import (
	"bytes"
	"regexp"
	"strings"
	"testing"

	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/testutil"
	"github.com/spf13/cobra"
)

var idPattern = regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)

func TestConnectorCustomizerCRUD(t *testing.T) {
	testutil.RequireLiveCredentials(t)

	spClient := requireConnectorClient(t)
	name := testutil.UniqueName("customizer")
	updatedName := name + "-updated"

	createCmd := newCustomizerCreateCmd(spClient)
	createOut := new(bytes.Buffer)
	createCmd.SetOut(createOut)
	createCmd.SetArgs([]string{name})

	if err := createCmd.Execute(); err != nil {
		testutil.SkipIfFeatureUnavailable(t, err)
		t.Fatalf("customizer create failed: %v", err)
	}

	customizerID := firstID(t, createOut.String())
	defer deleteCustomizer(t, spClient, customizerID)

	getCustomizerAndAssert(t, spClient, customizerID, name)

	updateCmd := newCustomizerUpdateCmd(spClient)
	updateOut := new(bytes.Buffer)
	updateCmd.SetOut(updateOut)
	updateCmd.Flags().Set("id", customizerID)
	updateCmd.Flags().Set("name", updatedName)

	if err := updateCmd.Execute(); err != nil {
		t.Fatalf("customizer update failed: %v", err)
	}
	if !strings.Contains(updateOut.String(), updatedName) {
		t.Fatalf("expected updated customizer output to contain %q, got %q", updatedName, updateOut.String())
	}

	getCustomizerAndAssert(t, spClient, customizerID, updatedName)
}

func TestConnectorCRUD(t *testing.T) {
	testutil.RequireLiveCredentials(t)

	spClient := requireConnectorClient(t)
	alias := testutil.UniqueName("connector")
	updatedAlias := alias + "-updated"

	createCmd := newConnCreateCmd(spClient)
	createOut := new(bytes.Buffer)
	createCmd.SetOut(createOut)
	createCmd.SetArgs([]string{alias})
	addConnEndpointFlag(createCmd)

	if err := createCmd.Execute(); err != nil {
		testutil.SkipIfFeatureUnavailable(t, err)
		t.Fatalf("connector create failed: %v", err)
	}

	connectorID := firstID(t, createOut.String())
	defer deleteConnector(t, spClient, connectorID)

	getConnectorAndAssert(t, spClient, connectorID, alias)

	updateCmd := newConnUpdateCmd(spClient)
	updateOut := new(bytes.Buffer)
	updateCmd.SetOut(updateOut)
	updateCmd.Flags().Set("id", connectorID)
	updateCmd.Flags().Set("alias", updatedAlias)
	addConnEndpointFlag(updateCmd)

	if err := updateCmd.Execute(); err != nil {
		t.Fatalf("connector update failed: %v", err)
	}
	if !strings.Contains(updateOut.String(), updatedAlias) {
		t.Fatalf("expected updated connector output to contain %q, got %q", updatedAlias, updateOut.String())
	}

	getConnectorAndAssert(t, spClient, connectorID, updatedAlias)
}

func requireConnectorClient(t *testing.T) client.Client {
	t.Helper()

	cfg, err := config.GetConfig()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	return client.NewSpClient(cfg)
}

func firstID(t *testing.T, output string) string {
	t.Helper()

	match := idPattern.FindString(output)
	if match == "" {
		t.Fatalf("expected UUID-like ID in output %q", output)
	}
	return match
}

func getCustomizerAndAssert(t *testing.T, spClient client.Client, id string, expectedName string) {
	t.Helper()

	getCmd := newCustomizerGetCmd(spClient)
	getOut := new(bytes.Buffer)
	getCmd.SetOut(getOut)
	getCmd.Flags().Set("id", id)
	if err := getCmd.Execute(); err != nil {
		t.Fatalf("customizer get failed: %v", err)
	}
	if !strings.Contains(getOut.String(), id) || !strings.Contains(getOut.String(), expectedName) {
		t.Fatalf("expected customizer get output to contain %q and %q, got %q", id, expectedName, getOut.String())
	}
}

func getConnectorAndAssert(t *testing.T, spClient client.Client, id string, expectedAlias string) {
	t.Helper()

	getCmd := newConnGetCmd(spClient)
	getOut := new(bytes.Buffer)
	getCmd.SetOut(getOut)
	getCmd.Flags().Set("id", id)
	addConnEndpointFlag(getCmd)
	if err := getCmd.Execute(); err != nil {
		t.Fatalf("connector get failed: %v", err)
	}
	if !strings.Contains(getOut.String(), id) || !strings.Contains(getOut.String(), expectedAlias) {
		t.Fatalf("expected connector get output to contain %q and %q, got %q", id, expectedAlias, getOut.String())
	}
}

func deleteCustomizer(t *testing.T, spClient client.Client, id string) {
	t.Helper()

	deleteCmd := newCustomizerDeleteCmd(spClient)
	deleteCmd.Flags().Set("id", id)
	if err := deleteCmd.Execute(); err != nil {
		t.Logf("failed to clean up connector customizer %s: %v", id, err)
	}
}

func deleteConnector(t *testing.T, spClient client.Client, id string) {
	t.Helper()

	deleteCmd := newConnDeleteCmd(spClient)
	deleteCmd.Flags().Set("id", id)
	addConnEndpointFlag(deleteCmd)
	if err := deleteCmd.Execute(); err != nil {
		t.Logf("failed to clean up connector %s: %v", id, err)
	}
}

func addConnEndpointFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP("conn-endpoint", "e", connectorsEndpoint, "Override connectors endpoint")
}
