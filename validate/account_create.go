package validate

import (
	"context"
	"time"

	"github.com/kr/pretty"
	"github.com/sailpoint/sp-cli/client"
)

var accountCreateChecks = []Check{
	{
		ID:             "account-create-empty",
		Description:    "Creating an account with no attributes should fail",
		IsDataModifier: true,
		RequiredCommands: []string{
			"std:account:create",
		},
		Run: func(ctx context.Context, spec *client.ConnSpec, cc *client.ConnClient, res *CheckResult) {
			input := map[string]interface{}{}
			_, _, err := cc.AccountCreate(ctx, nil, input)
			if err == nil {
				res.errf("expected error for empty account created")
			}
		},
	},
	{
		ID:             "account-create-minimal",
		Description:    "Creating an account with only required fields should be successful",
		IsDataModifier: true,
		RequiredCommands: []string{
			"std:account:create",
			"std:account:read",
			"std:account:delete",
		},
		Run: func(ctx context.Context, spec *client.ConnSpec, cc *client.ConnClient, res *CheckResult) {
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

			diffs := compareIntersection(input, acct.Attributes)
			for _, diff := range diffs {
				res.errf("input vs read mismatch %+v", diff)
			}

			acctRead, _, err := cc.AccountRead(ctx, acct.ID(), acct.UniqueID())
			if err != nil {
				res.errf("reading account: %s", err)
				return
			}
			diffs = compareIntersection(input, acctRead.Attributes)
			for _, diff := range diffs {
				res.errf("account diffs %+v", diff)
			}

			_, err = cc.AccountDelete(ctx, acct.ID(), acct.UniqueID())
			if err != nil {
				res.errf("deleting account: %s", err)
			}
		},
	},
	{
		ID:             "account-create-maximal",
		Description:    "Creating an account with all fields should be successful",
		IsDataModifier: true,
		RequiredCommands: []string{
			"std:account:create",
			"std:account:read",
			"std:account:delete",
		},
		Run: func(ctx context.Context, spec *client.ConnSpec, cc *client.ConnClient, res *CheckResult) {
			input := map[string]interface{}{}
			for _, field := range spec.AccountCreateTemplate.Fields {
				input[getFieldName(field)] = genCreateField(field)
			}

			identity := getIdentity(input)

			acct, _, err := cc.AccountCreate(ctx, &identity, input)
			if err != nil {
				res.errf("creating account: %s", err)
				return
			}

			diffs := compareIntersection(input, acct.Attributes)
			for _, diff := range diffs {
				res.errf("account diffs %+v", diff)
			}

			acctRead, _, err := cc.AccountRead(ctx, acct.ID(), acct.UniqueID())
			if err != nil {
				res.errf("reading account: %s", err)
				return
			}
			diffs = compareIntersection(input, acctRead.Attributes)
			for _, diff := range diffs {
				res.errf("account diffs %+v", diff)
			}

			_, err = cc.AccountDelete(ctx, acct.ID(), acct.UniqueID())
			if err != nil {
				res.errf("deleting account: %s", err)
			}
		},
	},
	{
		ID:             "account-create-list-delete",
		Description:    "Created accounts should show up in list accounts response; after deletion they should not",
		IsDataModifier: true,
		RequiredCommands: []string{
			"std:account:create",
			"std:account:read",
			"std:account:delete",
			"std:account:list",
		},
		Run: func(ctx context.Context, spec *client.ConnSpec, cc *client.ConnClient, res *CheckResult) {
			accountsPreCreate, _, err := cc.AccountList(ctx)
			if err != nil {
				res.err(err)
				return
			}

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

			accountsPostCreate, _, err := cc.AccountList(ctx)
			if err != nil {
				res.err(err)
				return
			}

			accountRead, _, err := cc.AccountRead(ctx, acct.ID(), acct.UniqueID())
			if err != nil {
				res.err(err)
				return
			}

			acctDiffs := pretty.Diff(*acct, *accountRead)
			if len(acctDiffs) > 0 {
				for _, diff := range acctDiffs {
					res.errf("[identity=%s] Diff: %s", acct.Identity, diff)
				}
			}

			_, err = cc.AccountDelete(ctx, acct.ID(), acct.UniqueID())
			if err != nil {
				res.errf("deleting account: %s", err)
			}

			// Allow deletion to propagate
			time.Sleep(5 * time.Second)

			_, _, err = cc.AccountRead(ctx, acct.ID(), acct.UniqueID())
			if err == nil {
				res.errf("was able to read deleted account: %q", acct.Identity)
			}

			accountsPostDelete, _, err := cc.AccountList(ctx)
			if err != nil {
				res.err(err)
			}

			if len(accountsPreCreate) != len(accountsPostDelete) {
				res.errf("expected # of accounts to match before creation (%d) and after deletion (%d)", len(accountsPreCreate), len(accountsPostDelete))
			}

			if len(accountsPreCreate)+1 != len(accountsPostCreate) {
				res.errf("expected # of accounts to be 1 larger after creation (%d) compare to before creation (%d)", len(accountsPostCreate), len(accountsPreCreate))
			}

			return
		},
	},
}
