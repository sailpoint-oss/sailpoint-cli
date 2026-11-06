// Copyright (c) 2026, SailPoint Technologies, Inc. All rights reserved.
package connector

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sailpoint-oss/sailpoint-cli/internal/mocks"
)

func okResponse() *http.Response {
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(nil))}
}

func TestExperimentalClientAddsHeaderToEveryMethod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	want := map[string]string{experimentalHeader: "true"}
	m := mocks.NewMockClient(ctrl)
	m.EXPECT().Get(gomock.Any(), "/get", gomock.Eq(want)).Return(okResponse(), nil)
	m.EXPECT().Delete(gomock.Any(), "/delete", gomock.Nil(), gomock.Eq(want)).Return(okResponse(), nil)
	m.EXPECT().Post(gomock.Any(), "/post", "application/json", gomock.Any(), gomock.Eq(want)).Return(okResponse(), nil)
	m.EXPECT().Put(gomock.Any(), "/put", "application/json", gomock.Any(), gomock.Eq(want)).Return(okResponse(), nil)
	m.EXPECT().Patch(gomock.Any(), "/patch", gomock.Any(), gomock.Eq(want)).Return(okResponse(), nil)

	c := newExperimentalClient(m)
	ctx := context.Background()
	_, _ = c.Get(ctx, "/get", nil)
	_, _ = c.Delete(ctx, "/delete", nil, nil)
	_, _ = c.Post(ctx, "/post", "application/json", nil, nil)
	_, _ = c.Put(ctx, "/put", "application/json", nil, nil)
	_, _ = c.Patch(ctx, "/patch", nil, nil)
}

func TestWithExperimentalHeaderKeepsCallerHeaders(t *testing.T) {
	in := map[string]string{"Accept": "application/json", experimentalHeader: "false"}
	got := withExperimentalHeader(in)

	if got["Accept"] != "application/json" {
		t.Fatalf("caller header dropped, got: %v", got)
	}
	if got[experimentalHeader] != "false" {
		t.Fatalf("caller-set %s was overwritten, got: %v", experimentalHeader, got)
	}
	if len(in) != 2 {
		t.Fatalf("input map was mutated: %v", in)
	}
}
