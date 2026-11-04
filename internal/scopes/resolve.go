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
	StatusScoped  = "scoped"
	StatusNoScope = "no-scope"
	// StatusFallback is an operation that lists no scope for the auth type:
	// the auth type is null or an empty list, and the operation is not
	// public. It gets FallbackScope, and the call might not work.
	StatusFallback = "fallback"
)

// FallbackScope is used when the API spec lists no scope for an operation.
const FallbackScope = "sp:scopes:all"

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
	// byAuth keeps the scope list from the spec, to tell null from [].
	byAuth := map[string][]string{}
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
			case c.Operation.Public:
				r.Status = StatusNoScope
			case len(alts) == 0:
				r.Status = StatusFallback
				r.Alternatives = []string{FallbackScope}
			default:
				r.Status = StatusScoped
			}
			byID[c.OperationID] = r
			byAuth[c.OperationID] = alts
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
		if r.Status == StatusNoScope {
			continue
		}
		r.Chosen = pick(r.Alternatives, chosen)
		if !chosen[r.Chosen] {
			chosen[r.Chosen] = true
			required = append(required, r.Chosen)
		}
		if r.Status == StatusFallback {
			warnings = append(warnings, fallbackWarning(id, auth, byAuth[id]))
		} else if r.DefaultScope {
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

func fallbackWarning(id string, auth AuthType, alts []string) string {
	reason := fmt.Sprintf("does not allow %s auth (%sAuth is null)", auth, auth)
	if alts != nil {
		reason = fmt.Sprintf("lists no scope for %s auth (%sAuth is [])", auth, auth)
	}
	token := "a user token"
	if auth == AuthApplication {
		token = "an application token"
	}
	return fmt.Sprintf("%s %s in the API spec. %s was added, but calls with %s might not work", id, reason, FallbackScope, token)
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
