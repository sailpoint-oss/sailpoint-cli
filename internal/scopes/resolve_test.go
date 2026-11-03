package scopes

import (
	"reflect"
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
	}

	ops, required, warnings := Resolve(calls, AuthUser)
	if !reflect.DeepEqual(required, []string{"idn:identity:read", "sp:scopes:all"}) {
		t.Errorf("user required = %v", required)
	}
	if ops[0].OperationID != "publicV1" || ops[0].Status != StatusNoScope {
		t.Errorf("public result = %+v", ops[0])
	}
	if len(warnings) != 1 {
		t.Errorf("user warnings = %v, want one default-scope warning", warnings)
	}

	ops, required, warnings = Resolve(calls, AuthApplication)
	if len(required) != 0 {
		t.Errorf("application required = %v, want none", required)
	}
	if ops[2].Status != StatusUnsupported {
		t.Errorf("userOnlyV1 status = %q, want unsupported", ops[2].Status)
	}
	if len(warnings) != 2 {
		t.Errorf("application warnings = %v, want two unsupported warnings", warnings)
	}
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
