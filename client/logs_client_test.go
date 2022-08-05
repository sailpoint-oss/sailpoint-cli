// Copyright (c) 2022, SailPoint Technologies, Inc. All rights reserved.
package client

import (
	"testing"
)

func TestLogFormatWithInnerLevel(t *testing.T) {
	msg := &LogMessage{
		TenantID: "123",
		Level:    "INFO",
		Message: map[string]interface{}{
			"level":   "DEBUG",
			"message": "log message",
		},
	}

	msgJson := msg.MessageString()
	if msgJson != "{\"message\":\"log message\"}" {
		t.Errorf("invalid format for json log message. expecting %s, got %s ", "{\"message\":\"log message\"}", msgJson)
	}

	msg = &LogMessage{
		TenantID: "123",
		Level:    "INFO",
		Message: map[string]interface{}{
			"message": "log message",
		},
	}

	msgJson = msg.MessageString()
	if msgJson != "{\"message\":\"log message\"}" {
		t.Errorf("invalid format for json log message. expecting %s, got %s ", "{\"message\":\"log message\"}", msgJson)
	}
}
