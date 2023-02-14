// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package connclient

import (
	"fmt"
	"testing"
)

const rawAccount = `{"identity":"john.doe","uuid":"1","attributes":{"email":"john.doe@example.com","firstName":"john","lastName":"doe"}}`

func TestParseDeprecatedAccountFormat(t *testing.T) {
	response := []byte(rawAccount)
	rawResps, state, err := parseResponse(response)
	if err != nil {
		t.Errorf("failed to parse account in deprecated format: %v", err)
	}

	if len(rawResps) != 1 {
		t.Errorf("deprecated account parsing error. expecting %d account, got %d ", 1, len(rawResps))
	}

	if state != nil {
		t.Errorf("does not expect state from deprecated account format, got %s", string(state.Data[:]))
	}
}

func TestParseNewAccountFormat(t *testing.T) {
	response := []byte(fmt.Sprintf(`{"type": "output", "data": %s}`, rawAccount))
	rawResps, state, err := parseResponse(response)
	if err != nil {
		t.Errorf("failed to parse account in deprecated format: %v", err)
	}

	if len(rawResps) != 1 {
		t.Errorf("deprecated account parsing error. expecting %d account, got %d ", 1, len(rawResps))
	}

	if state != nil {
		t.Errorf("does not expect state from deprecated account format, got %s", string(state.Data[:]))
	}
}

func TestParseNewAccountFormatWithState(t *testing.T) {
	stateStr := `{"foo": "bar"}`
	response := []byte(fmt.Sprintf(`{"type": "output", "data": %s}{"type": "state", "data": %s}`, rawAccount, stateStr))
	// accounts, state, printableResp, err := responseToAccounts(response)
	rawResps, state, err := parseResponse(response)
	if err != nil {
		t.Errorf("failed to parse account in deprecated format: %v", err)
	}

	if len(rawResps) != 1 {
		t.Errorf("deprecated account parsing error. expecting %d account, got %d ", 1, len(rawResps))
	}

	if state == nil {
		t.Errorf("failed to read state")
	}

	if string(state.Data[:]) != stateStr {
		t.Errorf("state is not in correct format. expecting %s, got %s", stateStr, string(state.Data[:]))
	}
}
