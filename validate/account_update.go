package validate

import (
	"context"
	"time"

	"github.com/sailpoint/sp-cli/client"
)

var accountUpdateChecks = []Check{
	{
		ID:             "account-update-single-attrs",
		Description:    "Test updating writable attributes",
		IsDataModifier: true,
		RequiredCommands: []string{
			"std:account:read",
			"std:account:list",
			"std:account:update",
		},
		Run: func(ctx context.Context, spec *client.ConnSpec, cc *client.ConnClient, res *CheckResult) {
			accounts, _, err := cc.AccountList(ctx)
			if err != nil {
				res.err(err)
			}

			if len(accounts) == 0 {
				res.warnf("account list is empty")
				return
			}

			acct := accounts[len(accounts)-1]

			for _, attr := range spec.AccountSchema.Attributes {
				if attr.Writable {
					if attr.Entitlement {
						// Skip entitlement field
						continue
					}
					change := attrChange(&acct, &attr)

					_, _, err = cc.AccountUpdate(ctx, acct.ID(), acct.UniqueID(), []client.AttributeChange{change})
					if err != nil {
						res.errf("update for %q failed: %s", attr.Name, err.Error())
						continue
					}

					// Give the update a chance to propagate
					time.Sleep(time.Second)

					acct, _, err := cc.AccountRead(ctx, acct.ID(), acct.UniqueID())
					if err != nil {
						res.err(err)
						continue
					}

					if acct.Attributes[attr.Name] != change.Value {
						res.errf("mismatch for %s. expected %+v; got %+v", attr.Name, change.Value, acct.Attributes[attr.Name])
						continue
					}
				}
			}
		},
	},
	{
		ID:             "account-update-entitlement",
		Description:    "Test updating entitlement field(s)",
		IsDataModifier: true,
		RequiredCommands: []string{
			"std:entitlement:list",
			"std:account:create",
			"std:account:read",
			"std:account:update",
			"std:account:delete",
		},
		Run: func(ctx context.Context, spec *client.ConnSpec, cc *client.ConnClient, res *CheckResult) {
			entitlementAttr := entitlementAttr(spec)
			if entitlementAttr == "" {
				res.warnf("no entitlement attribute")
				return
			}

			entitlements, _, err := cc.EntitlementList(ctx, "group")
			if err != nil {
				res.err(err)
				return
			}

			if len(entitlements) == 0 {
				res.warnf("no entitlements found")
				return
			}

			// Create minimal user
			input := map[string]interface{}{}
			for _, field := range spec.AccountCreateTemplate.Fields {
				if field.Required {
					input[getFieldName(field)] = genCreateField(field)
				}
			}

			identity := getIdentity(input)

			acct, _, err := cc.AccountCreate(ctx, &identity, input)
			if err != nil {
				res.errf("creating account: %s", err)
				return
			}

			// Give account creation a chance to propagate
			time.Sleep(time.Second)

			// Add entitlements
			for _, e := range entitlements {
				acct, _, err := cc.AccountRead(ctx, acct.ID(), acct.UniqueID())
				if err != nil {
					res.errf("failed to read account %q", acct.Identity)
				}

				accEntitlements, err := accountEntitlements(acct, spec)
				if err != nil {
					res.errf("failed to get acc entitlements")
				}

				if isAvailableForUpdating(accEntitlements, e.ID()) {
					_, _, err = cc.AccountUpdate(ctx, acct.ID(), acct.UniqueID(), []client.AttributeChange{
						{
							Op:        "Add",
							Attribute: entitlementAttr,
							Value:     e.ID(),
						},
					})
					if err != nil {
						res.errf("failed to add entitlement %q", e.Identity)
					}

					acct, _, err = cc.AccountRead(ctx, acct.ID(), acct.UniqueID())
					if err != nil {
						res.errf("failed to read account %q", acct.Identity)
					}

					if !accountHasEntitlement(acct, spec, e.ID()) {
						res.errf("failed to add entitlement: %q", e.ID())
					}
				}
			}

			// Remove entitlements
			for _, e := range entitlements {
				accEntitlements, err := accountEntitlements(acct, spec)
				if err != nil {
					res.errf("failed to get acc entitlements")
				}
				if len(accEntitlements) != 1 {
					_, _, err = cc.AccountUpdate(ctx, acct.ID(), acct.UniqueID(), []client.AttributeChange{
						{
							Op:        "Remove",
							Attribute: entitlementAttr,
							Value:     e.ID(),
						},
					})
					if err != nil {
						res.errf("failed to remove entitlement %q", e.ID())
					}

					acct, _, err := cc.AccountRead(ctx, acct.ID(), acct.UniqueID())
					if err != nil {
						res.errf("failed to read account %q", acct.ID())
					}

					if accountHasEntitlement(acct, spec, e.ID()) {
						res.errf("failed to remove entitlement: %q checking", e.ID())
					}
				}
			}

			_, err = cc.AccountDelete(ctx, acct.ID(), acct.UniqueID())
			if err != nil {
				res.errf("deleting account: %s", err)
			}
		},
	},
}
