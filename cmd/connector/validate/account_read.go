package connvalidate

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/kr/pretty"

	connclient "github.com/sailpoint-oss/sailpoint-cli/cmd/connector/client"
)

var accountReadChecks = []Check{
	{
		ID:             "account-list-and-read",
		Description:    "List accounts and read each account individual; ensure responses are equivalent",
		IsDataModifier: false,
		RequiredCommands: []string{
			"std:account:read",
			"std:account:list",
		},
		Run: func(ctx context.Context, spec *connclient.ConnSpec, cc *connclient.ConnClient, res *CheckResult, readLimit int64) {
			schema := cc.BuildAccountSchema(spec)
			accounts, _, _, err := cc.AccountList(ctx, nil, nil, schema)
			if err != nil {
				res.err(err)
				return
			}
			if len(accounts) == 0 {
				res.warnf("no accounts")
				return
			}

			rand.Shuffle(len(accounts), func(i, j int) {
				accounts[i], accounts[j] = accounts[j], accounts[i]
			})

			for index, account := range accounts {
				if int64(index) == readLimit {
					break
				}
				acct, _, err := cc.AccountRead(ctx, account.ID(), account.UniqueID(), schema)
				if err != nil {
					res.err(err)
					return
				}
				if acct.Identity != account.Identity {
					res.errf("want %q; got %q", account.Identity, acct.Identity)
				}

				canonicalizeAttributes(account.Attributes)
				canonicalizeAttributes(acct.Attributes)

				diffs := pretty.Diff(account, *acct)
				if len(diffs) > 0 {
					for _, diff := range diffs {
						res.errf("[identity=%s] Diff: %s", acct.Identity, diff)
					}
				}
			}
		},
	},
	{
		ID:             "account-not-found",
		Description:    "Reading an account based on an id which doesn't exist should fail",
		IsDataModifier: false,
		RequiredCommands: []string{
			"std:account:read",
		},
		Run: func(ctx context.Context, spec *connclient.ConnSpec, cc *connclient.ConnClient, res *CheckResult, readLimit int64) {
			schema := cc.BuildAccountSchema(spec)

			_, _, err := cc.AccountRead(ctx, "__sailpoint__not__found__", "", schema)
			if err == nil {
				res.errf("expected error for non-existant identity")
			}
		},
	},
	{
		ID:             "account-schema-check",
		Description:    "Verify account fields match schema",
		IsDataModifier: false,
		RequiredCommands: []string{
			"std:account:list",
		},
		Run: func(ctx context.Context, spec *connclient.ConnSpec, cc *connclient.ConnClient, res *CheckResult, readLimit int64) {
			additionalAttributes := map[string]string{}
			schema := cc.BuildAccountSchema(spec)

			attrsByName := map[string]connclient.AccountSchemaAttribute{}
			for _, value := range spec.AccountSchema.Attributes {
				attrsByName[value.Name] = value
			}

			accounts, _, _, err := cc.AccountList(ctx, nil, nil, schema)
			if err != nil {
				res.err(err)
				return
			}
			for _, acct := range accounts {
				for name, value := range acct.Attributes {
					attr, found := attrsByName[name]
					if !found {
						additionalAttributes[name] = ""
						continue
					}

					isMulti := false
					switch value.(type) {
					case []interface{}:
						if len(value.([]interface{})) > 0 {
							value = value.([]interface{})[0]
						} else {
							value = nil
						}
						isMulti = true
					}

					if attr.Multi != isMulti {
						res.errf("expected multi=%t but multi=%t", isMulti, attr.Multi)
					}

					switch value.(type) {
					case string:
						if attr.Type == "int" {
							_, err := strconv.Atoi(value.(string))
							if err != nil {
								res.errf("failed to convert int to string on field %s", name)
							}
						}

						if attr.Type != "string" && attr.Type != "int" {
							res.errf("expected type %q but was 'string'", attr.Type)
						}
					case bool:
						if attr.Type != "boolean" {
							res.errf("expected type %q but was 'boolean'", attr.Type)
						}
					case float64:
						if attr.Type != "int" {
							res.errf("expected type %q but was 'int'", attr.Type)
						}
					case nil:
						// okay
					default:
						panic(fmt.Sprintf("unknown type %T for %q", value, name))
					}
				}
			}

			for additional := range additionalAttributes {
				res.warnf("additional attribute %q", additional)
			}
		},
	},
}
