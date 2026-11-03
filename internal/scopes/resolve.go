package scopes

import (
	"fmt"
	"sort"
	"strings"
)

// AuthType selects which security scheme the scopes are read from.
type AuthType string

const (
	AuthUser        AuthType = "user"
	AuthApplication AuthType = "application"
)

// ParseAuthType validates a --auth-type value.
func ParseAuthType(s string) (AuthType, error) {
	switch AuthType(strings.ToLower(s)) {
	case AuthUser:
		return AuthUser, nil
	case AuthApplication:
		return AuthApplication, nil
	}
	return "", fmt.Errorf("invalid auth type %q (valid: user, application)", s)
}

// Operation status values reported in OperationResult.Status.
const (
	StatusScoped      = "scoped"
	StatusNoScope     = "no-scope"
	StatusUnsupported = "unsupported"
)

// OperationResult is the scope decision for one distinct operation.
type OperationResult struct {
	OperationID  string   `json:"operationId"`
	HTTPMethod   string   `json:"httpMethod"`
	Path         string   `json:"path"`
	Status       string   `json:"status"`
	Alternatives []string `json:"alternatives"`
	Chosen       string   `json:"chosen,omitempty"`
	DefaultScope bool     `json:"defaultScope,omitempty"`
	Calls        int      `json:"calls"`
}

// Result is the full outcome of an analysis.
type Result struct {
	AuthType       AuthType          `json:"authType"`
	Sources        []Source          `json:"sources"`
	Calls          []Call            `json:"calls"`
	Operations     []OperationResult `json:"operations"`
	RequiredScopes []string          `json:"requiredScopes"`
	Warnings       []string          `json:"warnings"`
}

// Resolve picks the smallest set of scopes that covers every call.
//
// Each operation accepts any one of its listed scopes. Operations with a
// single option are processed first, so that their scopes are chosen before
// the operations that have a choice. An operation with a choice uses a scope
// that is already chosen when it can. Otherwise it takes a read scope if one
// is listed, else the first scope in the list.
func Resolve(calls []Call, auth AuthType) ([]OperationResult, []string, []string) {
	byID := map[string]*OperationResult{}
	var ids []string
	for _, c := range calls {
		r, ok := byID[c.OperationID]
		if !ok {
			alts := c.Operation.ScopesFor(auth)
			r = &OperationResult{
				OperationID:  c.OperationID,
				HTTPMethod:   c.Operation.HTTPMethod,
				Path:         c.Operation.Path,
				Alternatives: append([]string{}, alts...),
				DefaultScope: c.Operation.DefaultScope,
			}
			switch {
			case alts == nil:
				r.Status = StatusUnsupported
			case len(alts) == 0:
				r.Status = StatusNoScope
			default:
				r.Status = StatusScoped
			}
			byID[c.OperationID] = r
			ids = append(ids, c.OperationID)
		}
		r.Calls++
	}

	order := append([]string{}, ids...)
	sort.SliceStable(order, func(i, j int) bool {
		a, b := byID[order[i]], byID[order[j]]
		if len(a.Alternatives) != len(b.Alternatives) {
			return len(a.Alternatives) < len(b.Alternatives)
		}
		return a.OperationID < b.OperationID
	})

	chosen := map[string]bool{}
	var required []string
	var warnings []string
	for _, id := range order {
		r := byID[id]
		switch r.Status {
		case StatusUnsupported:
			warnings = append(warnings, fmt.Sprintf("%s cannot be called with %s auth", id, auth))
			continue
		case StatusNoScope:
			continue
		}
		r.Chosen = pick(r.Alternatives, chosen)
		if !chosen[r.Chosen] {
			chosen[r.Chosen] = true
			required = append(required, r.Chosen)
		}
		if r.DefaultScope {
			warnings = append(warnings, fmt.Sprintf("%s documents no scope and falls back to %s", id, r.Chosen))
		}
	}

	results := make([]OperationResult, 0, len(ids))
	for _, id := range ids {
		results = append(results, *byID[id])
	}
	sort.Slice(results, func(i, j int) bool { return results[i].OperationID < results[j].OperationID })
	sort.Strings(required)
	sort.Strings(warnings)
	return results, required, warnings
}

func pick(alts []string, chosen map[string]bool) string {
	for _, a := range alts {
		if chosen[a] {
			return a
		}
	}
	for _, a := range alts {
		if strings.HasSuffix(a, ":read") {
			return a
		}
	}
	return alts[0]
}
