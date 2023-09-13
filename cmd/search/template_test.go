// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.

package search

import (
	"bytes"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
)

func TestNewTemplateCommand(t *testing.T) {
	err := config.InitConfig()
	if err != nil {
		t.Fatalf("Error initializing config: %v", err)
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmd := newTemplateCmd("", false)

	b := new(bytes.Buffer)
	cmd.SetOut(b)
	cmd.SetArgs([]string{"all-provisioning-events-90-days"})
	cmd.Flags().Set("wait", "true")

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("TestNewTemplateCommand: Unable to execute the command successfully: %v", err)
	}
}
