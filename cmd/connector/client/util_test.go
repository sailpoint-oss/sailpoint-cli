// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package connclient

import (
	"fmt"
	"testing"
)

const rawAccount = `{"identity":"john.doe","uuid":"1","attributes":{"email":"john.doe@example.com","firstName":"john","lastName":"doe"}}`

func TestParseDeprecatedAccountListFormat(t *testing.T) {
	response := []byte(rawAccount)
	rawResps, state, err := parseResponseList(response)
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

func TestParseNewAccountListFormat(t *testing.T) {
	response := []byte(fmt.Sprintf(`{"type": "output", "data": %s}`, rawAccount))
	rawResps, state, err := parseResponseList(response)
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

func TestParseNewAccountListFormatWithState(t *testing.T) {
	stateStr := `{"foo": "bar"}`
	response := []byte(fmt.Sprintf(`{"type": "output", "data": %s}{"type": "state", "data": %s}`, rawAccount, stateStr))
	// accounts, state, printableResp, err := responseToAccounts(response)
	rawResps, state, err := parseResponseList(response)
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

func TestParseDepracatedAccountFormat(t *testing.T) {
	response := []byte(rawAccount)
	rawResp, err := parseResponse(response)
	if err != nil {
		t.Errorf("failed to parse account in deprecated format: %v", err)
	}

	if string(rawResp.Data[:]) != rawAccount {
		t.Errorf("state is not in correct format. expecting %s, got %s", rawAccount, string(rawResp.Data[:]))
	}
}

func TestParseNewAccountFormat(t *testing.T) {
	response := []byte(fmt.Sprintf(`{"type": "output", "data": %s}`, rawAccount))
	rawResp, err := parseResponse(response)
	if err != nil {
		t.Errorf("failed to parse account in deprecated format: %v", err)
	}

	if string(rawResp.Data[:]) != rawAccount {
		t.Errorf("state is not in correct format. expecting %s, got %s", rawAccount, string(rawResp.Data[:]))
	}
}
