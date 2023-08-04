// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.

package transform

import (
	"bytes"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
)

func TestNewListCmd(t *testing.T) {

	err := config.InitConfig()
	if err != nil {
		t.Fatalf("Error initializing config: %v", err)
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmd := newListCommand()

	b := new(bytes.Buffer)
	cmd.SetOut(b)

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("error execute cmd: %v", err)
	}
}
