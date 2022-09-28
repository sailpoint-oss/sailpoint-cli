package connvalidate

import (
	"context"
	"encoding/json"

	connclient "github.com/sailpoint-oss/sailpoint-cli/cmd/connector/client"
)

var testConnChecks = []Check{
	{
		ID:             "test-connection-empty",
		Description:    "Verify that test connection fails with an empty config",
		IsDataModifier: false,
		RequiredCommands: []string{
			"std:test-connection",
		},
		Run: func(ctx context.Context, spec *connclient.ConnSpec, cc *connclient.ConnClient, res *CheckResult) {
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
		Run: func(ctx context.Context, spec *connclient.ConnSpec, cc *connclient.ConnClient, res *CheckResult) {
			_, err := cc.TestConnection(ctx)
			if err != nil {
				res.err(err)
			}
		},
	},
}
