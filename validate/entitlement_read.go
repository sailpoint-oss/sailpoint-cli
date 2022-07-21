package validate

import (
	"context"
	"github.com/kr/pretty"
	"github.com/sailpoint/sp-cli/client"
)

var entitlementReadChecks = []Check{
	{
		ID:             "entitlement-not-found",
		Description:    "Verify reading a non existant entitlement fails",
		IsDataModifier: false,
		RequiredCommands: []string{
			"std:entitlement:read",
		},
		Run: func(ctx context.Context, spec *client.ConnSpec, cc *client.ConnClient, res *CheckResult) {
			_, _, err := cc.EntitlementRead(ctx, "__sailpoint__not__found__", "", "group")
			if err == nil {
				res.errf("expected error for non-existant entitlement")
			}
			return
		},
	},
	{
		ID:             "entitlement-list-read",
		Description:    "Verify that we can list each entitlement and then read; results should match",
		IsDataModifier: false,
		RequiredCommands: []string{
			"std:entitlement:read",
			"std:entitlement:list",
		},
		Run: func(ctx context.Context, spec *client.ConnSpec, cc *client.ConnClient, res *CheckResult) {
			entitlements, _, err := cc.EntitlementList(ctx, "group")
			if err != nil {
				res.err(err)
				return
			}

			if len(entitlements) == 0 {
				res.warnf("no entitlements")
				return
			}

			for _, e := range entitlements {
				eRead, _, err := cc.EntitlementRead(ctx, e.ID(), e.UniqueID(), "group")
				if err != nil {
					res.errf("failed to read entitlement %q: %s", e.Identity, err.Error())
					return
				}

				if e.Identity != eRead.Identity {
					res.errf("want %q; got %q", e.Identity, eRead.Identity)
				}
				diffs := pretty.Diff(e, *eRead)
				if len(diffs) > 0 {
					for _, diff := range diffs {
						res.errf("[identity=%s] Diff: %s", e.Identity, diff)
					}
				}
			}
		},
	},
	{
		ID:             "entitlement-schema-check",
		Description:    "Verify entitlement schema field match",
		IsDataModifier: false,
		RequiredCommands: []string{
			"std:entitlement:list",
		},
		Run: func(ctx context.Context, spec *client.ConnSpec, cc *client.ConnClient, res *CheckResult) {
			additionalAttributes := map[string]string{}

			attrsByName := map[string]client.EntitlementSchemaAttribute{}
			for _, value := range spec.EntitlementSchemas[0].Attributes {
				attrsByName[value.Name] = value
			}

			entitlements, _, err := cc.EntitlementList(ctx, "group")
			if err != nil {
				res.err(err)
				return
			}
			for _, acct := range entitlements {
				for name, value := range acct.Attributes {
					attr, found := attrsByName[name]
					if !found {
						additionalAttributes[name] = ""
						continue
					}

					testSchema(res, name, value, attr.Multi, attr.Type)
				}
			}

			for additional := range additionalAttributes {
				res.warnf("additional attribute %q", additional)
			}
		},
	},
}
