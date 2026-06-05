package env

import (
	"bytes"
	"testing"

	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/testutil"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestEnvLocalCRUD(t *testing.T) {
	testutil.SetJSONOutput(t)

	previousEnvironments := viper.GetStringMap("environments")
	previousActive := viper.GetString("activeenvironment")
	t.Cleanup(func() {
		viper.Set("environments", previousEnvironments)
		viper.Set("activeenvironment", previousActive)
	})

	envName := "sail-cli-ci-env"
	environments := map[string]any{
		envName: map[string]any{
			"tenanturl": "https://tenant.identitynow.com",
			"baseurl":   "https://tenant.api.identitynow.com",
			"authtype":  "pat",
		},
	}
	viper.Set("environments", environments)
	config.SetActiveEnvironment(envName)

	showOut := executeEnvCommand(t, newShowCommand(), []string{envName})
	shown := testutil.DecodeJSON[map[string]any](t, showOut)
	if shown["name"] != envName {
		t.Fatalf("expected env show name %q, got %#v", envName, shown["name"])
	}

	listOut := executeEnvCommand(t, newListCommand(), nil)
	listed := testutil.DecodeJSON[[]map[string]any](t, listOut)
	if len(listed) != 1 || listed[0]["name"] != envName {
		t.Fatalf("expected env list to contain %q, got %#v", envName, listed)
	}

	otherEnv := "sail-cli-ci-env-other"
	environments[otherEnv] = map[string]any{
		"tenanturl": "https://other.identitynow.com",
		"baseurl":   "https://other.api.identitynow.com",
		"authtype":  "oauth",
	}
	viper.Set("environments", environments)

	useCmd := newUseCommand()
	useCmd.SetArgs([]string{otherEnv})
	if err := useCmd.Execute(); err != nil {
		t.Fatalf("env use failed: %v", err)
	}
	if config.GetActiveEnvironment() != otherEnv {
		t.Fatalf("expected active environment %q, got %q", otherEnv, config.GetActiveEnvironment())
	}

	deleteCmd := newDeleteCommand()
	deleteCmd.SetArgs([]string{envName})
	deleteCmd.Flags().Set("force", "true")
	if err := deleteCmd.Execute(); err != nil {
		t.Fatalf("env delete failed: %v", err)
	}
	if config.GetEnvironments()[envName] != nil {
		t.Fatalf("expected environment %q to be deleted", envName)
	}
}

func executeEnvCommand(t *testing.T, cmd *cobra.Command, args []string) string {
	t.Helper()

	out := new(bytes.Buffer)
	cmd.SetOut(out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("env command failed: %v", err)
	}
	return out.String()
}
