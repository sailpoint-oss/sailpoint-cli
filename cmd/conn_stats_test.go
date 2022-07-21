// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package cmd

import (
	"testing"
	"time"
)

func Test_parseDuration(t *testing.T) {

	tests := []struct {
		name        string
		durationStr string
		want        time.Duration
		wantErr     bool
	}{
		{
			name:        "1. valid day",
			durationStr: "1d",
			want:        24 * time.Hour,
			wantErr:     false,
		},
		{
			name:        "2. valid 9 day",
			durationStr: "9d",
			want:        9 * 24 * time.Hour,
			wantErr:     false,
		},
		{
			name:        "3. Invalid 15 day",
			durationStr: "15d",
			wantErr:     true,
		},
		{
			name:        "4. valid week",
			durationStr: "1w",
			want:        7 * 24 * time.Hour,
			wantErr:     false,
		},
		{
			name:        "5. valid 9 week",
			durationStr: "9w",
			want:        9 * 7 * 24 * time.Hour,
			wantErr:     false,
		},
		{
			name:        "6. Invalid 15 week",
			durationStr: "15d",
			wantErr:     true,
		},
		{
			name:        "7. Invalid text",
			durationStr: "sd",
			wantErr:     true,
		},
		{
			name:        "8. Invalid text",
			durationStr: "sde",
			wantErr:     true,
		},
		{
			name:        "9. Invalid text",
			durationStr: "234",
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDuration(tt.durationStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTenantStats.getCommandStats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (got != nil && *got != tt.want) || (got == nil && err == nil) {
				t.Errorf("validDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}
