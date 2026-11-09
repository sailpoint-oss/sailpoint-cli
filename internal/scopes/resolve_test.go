package scopes

import (
	"reflect"
	"strings"
	"testing"
)

func call(id string, op Operation) Call {
	return Call{OperationID: id, Operation: op}
}

func TestResolvePrefersSharedScope(t *testing.T) {
	calls := []Call{
		call("listAccountsV1", Operation{UserAuth: []string{"idn:accounts:read", "idn:accounts:manage"}}),
		call("deleteAccountV1", Operation{UserAuth: []string{"idn:accounts:manage"}}),
		call("listAccountsV1", Operation{UserAuth: []string{"idn:accounts:read", "idn:accounts:manage"}}),
	}
	ops, required, warnings := Resolve(calls, AuthUser)
	if !reflect.DeepEqual(required, []string{"idn:accounts:manage"}) {
		t.Errorf("required = %v, want [idn:accounts:manage]", required)
	}
	if len(warnings) != 0 {
		t.Errorf("warnings = %v, want none", warnings)
	}
	if ops[1].OperationID != "listAccountsV1" || ops[1].Calls != 2 || ops[1].Chosen != "idn:accounts:manage" {
		t.Errorf("listAccountsV1 result = %+v", ops[1])
	}
}

func TestResolvePrefersReadScope(t *testing.T) {
	calls := []Call{call("getSourceV1", Operation{UserAuth: []string{"idn:sources:manage", "idn:sources:read"}})}
	_, required, _ := Resolve(calls, AuthUser)
	if !reflect.DeepEqual(required, []string{"idn:sources:read"}) {
		t.Errorf("required = %v, want [idn:sources:read]", required)
	}
}

func TestResolveSpecialCases(t *testing.T) {
	calls := []Call{
		call("publicV1", Operation{UserAuth: []string{}, ApplicationAuth: []string{}, Public: true}),
		call("undocumentedV1", Operation{UserAuth: []string{"sp:scopes:all"}, DefaultScope: true}),
		call("userOnlyV1", Operation{UserAuth: []string{"idn:identity:read"}}),
		call("noScopeV1", Operation{UserAuth: []string{}}),
	}

	ops, required, warnings := Resolve(calls, AuthUser)
	if !reflect.DeepEqual(required, []string{"idn:identity:read", "sp:scopes:all"}) {
		t.Errorf("user required = %v", required)
	}
	if ops[1].OperationID != "publicV1" || ops[1].Status != StatusNoScope {
		t.Errorf("public result = %+v", ops[1])
	}
	if ops[0].OperationID != "noScopeV1" || ops[0].Status != StatusFallback || ops[0].Chosen != FallbackScope {
		t.Errorf("empty userAuth result = %+v", ops[0])
	}
	if len(warnings) != 2 {
		t.Errorf("user warnings = %v, want default-scope and fallback warnings", warnings)
	}
	if !containsText(warnings, "noScopeV1 lists no scope for user auth (userAuth is [])") {
		t.Errorf("user warnings = %v, want an empty userAuth warning", warnings)
	}

	ops, required, warnings = Resolve(calls, AuthApplication)
	if !reflect.DeepEqual(required, []string{"sp:scopes:all"}) {
		t.Errorf("application required = %v, want [sp:scopes:all]", required)
	}
	if ops[3].Status != StatusFallback {
		t.Errorf("userOnlyV1 status = %q, want fallback", ops[3].Status)
	}
	if !containsText(warnings, "userOnlyV1 does not allow application auth (applicationAuth is null)") ||
		!containsText(warnings, "might not work") {
		t.Errorf("application warnings = %v", warnings)
	}
}

func containsText(list []string, text string) bool {
	for _, s := range list {
		if strings.Contains(s, text) {
			return true
		}
	}
	return false
}

func TestSnapshotLoads(t *testing.T) {
	m, err := Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	op, ok := m.Operations["listEntitlementsByIdentityV1"]
	if !ok {
		t.Fatal("snapshot has no listEntitlementsByIdentityV1")
	}
	if !reflect.DeepEqual(op.UserAuth, []string{"idn:entitlement:read"}) {
		t.Errorf("userAuth = %v, want [idn:entitlement:read]", op.UserAuth)
	}
	if got := m.NameIndex(LangPowerShell)["get-entitlementsbyidentityv1"]; got != "listEntitlementsByIdentityV1" {
		t.Errorf("powershell index = %q", got)
	}
}
