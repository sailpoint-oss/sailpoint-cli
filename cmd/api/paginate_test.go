// Copyright (c) 2024, SailPoint Technologies, Inc. All rights reserved.
package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sailpoint-oss/sailpoint-cli/internal/mocks"
)

func TestParseQueryParams(t *testing.T) {
	tests := []struct {
		name      string
		params    []string
		wantLimit int
		wantOff   int
		hasLimit  bool
		hasOffset bool
	}{
		{
			name:   "empty",
			params: nil,
		},
		{
			name:      "limit only",
			params:    []string{"limit=100"},
			wantLimit: 100,
			hasLimit:  true,
		},
		{
			name:      "offset only",
			params:    []string{"offset=500"},
			wantOff:   500,
			hasOffset: true,
		},
		{
			name:      "both",
			params:    []string{"limit=50", "offset=200"},
			wantLimit: 50,
			wantOff:   200,
			hasLimit:  true,
			hasOffset: true,
		},
		{
			name:   "other params ignored",
			params: []string{"filters=name eq \"test\"", "sorters=name"},
		},
		{
			name:      "mixed with other params",
			params:    []string{"filters=name eq \"test\"", "limit=10", "offset=20"},
			wantLimit: 10,
			wantOff:   20,
			hasLimit:  true,
			hasOffset: true,
		},
		{
			name:   "invalid limit ignored",
			params: []string{"limit=abc"},
		},
		{
			name:   "malformed param ignored",
			params: []string{"noequalssign"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limit, offset, hasLimit, hasOffset := parseQueryParams(tt.params)
			if limit != tt.wantLimit {
				t.Errorf("limit = %d, want %d", limit, tt.wantLimit)
			}
			if offset != tt.wantOff {
				t.Errorf("offset = %d, want %d", offset, tt.wantOff)
			}
			if hasLimit != tt.hasLimit {
				t.Errorf("hasLimit = %v, want %v", hasLimit, tt.hasLimit)
			}
			if hasOffset != tt.hasOffset {
				t.Errorf("hasOffset = %v, want %v", hasOffset, tt.hasOffset)
			}
		})
	}
}

func TestBuildPaginatedEndpoint(t *testing.T) {
	tests := []struct {
		name         string
		endpoint     string
		queryParams  []string
		offset       int
		limit        int
		includeCount bool
		wantContains []string
		wantExcludes []string
	}{
		{
			name:         "basic with count",
			endpoint:     "/v2024/accounts",
			offset:       0,
			limit:        250,
			includeCount: true,
			wantContains: []string{"limit=250", "offset=0", "count=true"},
		},
		{
			name:         "without count",
			endpoint:     "/v2024/accounts",
			offset:       250,
			limit:        250,
			includeCount: false,
			wantContains: []string{"limit=250", "offset=250"},
			wantExcludes: []string{"count="},
		},
		{
			name:         "user params preserved, limit/offset/count stripped",
			endpoint:     "/v2024/accounts",
			queryParams:  []string{"filters=name eq \"test\"", "limit=100", "offset=50", "count=false"},
			offset:       0,
			limit:        100,
			includeCount: true,
			wantContains: []string{"limit=100", "offset=0", "count=true", "filters="},
		},
		{
			name:         "no duplicate user params",
			endpoint:     "/v2024/accounts",
			queryParams:  []string{"sorters=name"},
			offset:       500,
			limit:        50,
			includeCount: false,
			wantContains: []string{"limit=50", "offset=500", "sorters=name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := buildPaginatedEndpoint(tt.endpoint, tt.queryParams, tt.offset, tt.limit, tt.includeCount)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, s := range tt.wantContains {
				if !strings.Contains(result, s) {
					t.Errorf("result %q does not contain %q", result, s)
				}
			}
			for _, s := range tt.wantExcludes {
				if strings.Contains(result, s) {
					t.Errorf("result %q should not contain %q", result, s)
				}
			}
		})
	}
}

func makeResponse(statusCode int, status string, body string, headers map[string]string) *http.Response {
	h := http.Header{}
	for k, v := range headers {
		h.Set(k, v)
	}
	return &http.Response{
		StatusCode: statusCode,
		Status:     status,
		Header:     h,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestPaginatedGet_MergesArrays(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClient(ctrl)

	gomock.InOrder(
		mockClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(
			makeResponse(200, "200 OK", `[{"id":1},{"id":2},{"id":3}]`, map[string]string{"X-Total-Count": "7"}), nil),
		mockClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(
			makeResponse(200, "200 OK", `[{"id":4},{"id":5},{"id":6}]`, nil), nil),
		mockClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(
			makeResponse(200, "200 OK", `[{"id":7}]`, nil), nil),
	)

	pageCfg := PaginationConfig{All: true, Limit: 3, Offset: 0}
	body, status, err := paginatedGet(context.Background(), mockClient, "/v2024/accounts", map[string]string{}, nil, pageCfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var items []json.RawMessage
	if err := json.Unmarshal(body, &items); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(items) != 7 {
		t.Errorf("expected 7 items, got %d", len(items))
	}

	if !strings.Contains(status, "7 items") {
		t.Errorf("expected status to mention 7 items, got: %s", status)
	}
}

func TestPaginatedGet_NonArrayResponse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClient(ctrl)

	mockClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(
		makeResponse(200, "200 OK", `{"id":1,"name":"test"}`, map[string]string{"X-Total-Count": "1"}), nil)

	pageCfg := PaginationConfig{All: true, Limit: 250, Offset: 0}
	_, _, err := paginatedGet(context.Background(), mockClient, "/v2024/accounts/123", map[string]string{}, nil, pageCfg)
	if err == nil {
		t.Fatal("expected error for non-array response")
	}
	if !strings.Contains(err.Error(), "pagination only supported for endpoints returning JSON arrays") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPaginatedGet_StopsOnShortPage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClient(ctrl)

	// Page 1 returns only 2 items with limit=5, so pagination stops immediately
	mockClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(
		makeResponse(200, "200 OK", `[{"id":1},{"id":2}]`, map[string]string{"X-Total-Count": "2"}), nil)

	pageCfg := PaginationConfig{Pages: 3, Limit: 5, Offset: 0}
	body, _, err := paginatedGet(context.Background(), mockClient, "/v2024/accounts", map[string]string{}, nil, pageCfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var items []json.RawMessage
	if err := json.Unmarshal(body, &items); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestPaginatedGet_AllMode_MissingTotalCount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClient(ctrl)

	mockClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(
		makeResponse(200, "200 OK", `[{"id":1}]`, nil), nil)

	pageCfg := PaginationConfig{All: true, Limit: 250, Offset: 0}
	_, _, err := paginatedGet(context.Background(), mockClient, "/v2024/accounts", map[string]string{}, nil, pageCfg)
	if err == nil {
		t.Fatal("expected error when X-Total-Count is missing with --all")
	}
	if !strings.Contains(err.Error(), "X-Total-Count") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPaginatedGet_PagesMode_MissingTotalCount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClient(ctrl)

	gomock.InOrder(
		mockClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(
			makeResponse(200, "200 OK", `[{"id":1},{"id":2},{"id":3}]`, nil), nil),
		mockClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(
			makeResponse(200, "200 OK", `[{"id":4}]`, nil), nil),
	)

	pageCfg := PaginationConfig{Pages: 2, Limit: 3, Offset: 0}
	body, _, err := paginatedGet(context.Background(), mockClient, "/v2024/accounts", map[string]string{}, nil, pageCfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var items []json.RawMessage
	if err := json.Unmarshal(body, &items); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(items) != 4 {
		t.Errorf("expected 4 items, got %d", len(items))
	}
}

func TestPaginatedGet_RejectsInvalidLimitAndOffset(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClient(ctrl)

	tests := []struct {
		name string
		cfg  PaginationConfig
		want string
	}{
		{
			name: "zero limit",
			cfg:  PaginationConfig{Pages: 2, Limit: 0, Offset: 0},
			want: "limit must be greater than 0",
		},
		{
			name: "negative limit",
			cfg:  PaginationConfig{Pages: 2, Limit: -1, Offset: 0},
			want: "limit must be greater than 0",
		},
		{
			name: "negative offset",
			cfg:  PaginationConfig{Pages: 2, Limit: 250, Offset: -1},
			want: "offset must be greater than or equal to 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := paginatedGet(context.Background(), mockClient, "/v2024/accounts", map[string]string{}, nil, tt.cfg)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected %q in error, got %v", tt.want, err)
			}
		})
	}
}

func TestPaginatedGet_AllModeDoesNotTrustShortFirstPageOverTotalCount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClient(ctrl)

	gomock.InOrder(
		mockClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(
			makeResponse(200, "200 OK", `[{"id":1},{"id":2}]`, map[string]string{"X-Total-Count": "7"}), nil),
		mockClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(
			makeResponse(200, "200 OK", `[{"id":6},{"id":7}]`, nil), nil),
	)

	pageCfg := PaginationConfig{All: true, Limit: 5, Offset: 0}
	body, _, err := paginatedGet(context.Background(), mockClient, "/v2024/accounts", map[string]string{}, nil, pageCfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var items []json.RawMessage
	if err := json.Unmarshal(body, &items); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(items) != 4 {
		t.Errorf("expected items from both pages, got %d", len(items))
	}
}

func TestPaginatedGet_RespectsUserLimitOffset(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClient(ctrl)

	// With limit=2 and offset=10, the first request should use those values
	var capturedURL string
	mockClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, url string, headers map[string]string) (*http.Response, error) {
			capturedURL = url
			return makeResponse(200, "200 OK", `[{"id":1}]`, map[string]string{"X-Total-Count": "11"}), nil
		})

	pageCfg := PaginationConfig{Pages: 1, Limit: 2, Offset: 10}
	_, _, err := paginatedGet(context.Background(), mockClient, "/v2024/accounts", map[string]string{}, nil, pageCfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(capturedURL, "limit=2") {
		t.Errorf("expected URL to contain limit=2, got: %s", capturedURL)
	}
	if !strings.Contains(capturedURL, "offset=10") {
		t.Errorf("expected URL to contain offset=10, got: %s", capturedURL)
	}
}

func TestPaginatedGet_Retry429(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockClient(ctrl)

	gomock.InOrder(
		// First call returns 429 with Retry-After: 0 for instant retry
		mockClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(
			makeResponse(429, "429 Too Many Requests", "", map[string]string{"Retry-After": "0"}), nil),
		// Second call succeeds
		mockClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(
			makeResponse(200, "200 OK", `[{"id":1}]`, map[string]string{"X-Total-Count": "1"}), nil),
	)

	pageCfg := PaginationConfig{Pages: 1, Limit: 250, Offset: 0}
	body, _, err := paginatedGet(context.Background(), mockClient, "/v2024/accounts", map[string]string{}, nil, pageCfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var items []json.RawMessage
	if err := json.Unmarshal(body, &items); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}
}
