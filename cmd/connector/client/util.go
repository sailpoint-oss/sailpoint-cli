// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package connclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

const (
	ResponseTypeOutput = "output"
	ResponseTypeState  = "state"
)

// RawResponse represents the response format from the connector
type RawResponse struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// parseResponse parses a single response into the new format if the given response is in deprecated format
func parseResponse(resp []byte) (rawResp *RawResponse, err error) {
	rawResps, _, err := parseResponseList(resp)
	if err != nil {
		return nil, err
	}

	if len(rawResps) != 1 {
		return nil, fmt.Errorf("failed to parse response")
	}

	return &rawResps[0], nil
}

// parseResponseList parses responses into the new format if the given response is in deprecated format
func parseResponseList(resp []byte) (rawResps []RawResponse, state *RawResponse, err error) {
	decoder := json.NewDecoder(bytes.NewReader(resp))
	deprecatedFormat := false
	for {
		rr := &RawResponse{}
		err = decoder.Decode(rr)
		if err != nil {
			if err == io.EOF {
				err = nil
				break
			}
			return nil, nil, err
		}

		if rr.Type == "" || rr.Data == nil {
			deprecatedFormat = true
			break
		}

		if rr.Type == ResponseTypeOutput {
			rawResps = append(rawResps, *rr)

		}

		if rr.Type == ResponseTypeState {
			state = rr
		}
	}

	if deprecatedFormat {
		decoder := json.NewDecoder(bytes.NewReader(resp))
		for {
			rr := json.RawMessage{}
			err = decoder.Decode(&rr)
			if err != nil {
				if err == io.EOF {
					err = nil
					break
				}
				return nil, nil, err
			}

			rawResps = append(rawResps, RawResponse{
				Type: ResponseTypeOutput,
				Data: rr,
			})

		}
	}

	return rawResps, state, err
}
