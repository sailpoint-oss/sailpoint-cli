package validate

import (
	"context"
	"encoding/json"

	"github.com/sailpoint/sp-cli/client"
)

var testConnChecks = []Check{
	{
		ID:             "test-connection-empty",
		Description:    "Verify that test connection fails with an empty config",
		IsDataModifier: false,
		RequiredCommands: []string{
			"std:test-connection",
		},
		Run: func(ctx context.Context, spec *client.ConnSpec, cc *client.ConnClient, res *CheckResult) {
			err := cc.TestConnectionWithConfig(ctx, json.RawMessage("{}"))
			if err == nil {
				res.errf("expected test-connection failure for empty config")
			}
		},
	},
	{
		ID:             "test-connection-success",
		Description:    "Verify that test connection succeeds with provided config",
		IsDataModifier: false,
		RequiredCommands: []string{
			"std:test-connection",
		},
		Run: func(ctx context.Context, spec *client.ConnSpec, cc *client.ConnClient, res *CheckResult) {
			_, err := cc.TestConnection(ctx)
			if err != nil {
				res.err(err)
			}
		},
	},
}
