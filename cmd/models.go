// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package cmd

import (
	"fmt"
	"strconv"
)

type connector struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Alias       string `json:"alias"`
}

func (c connector) columns() []string {
	return []string{c.ID, c.Alias}
}

var connectorColumns = []string{"ID", "Alias"}

type connectorVersion struct {
	ConnectorID string `json:"connectorId"`
	Version     int    `json:"version"`
}

type connectorUpdate struct {
	DisplayName string `json:"displayName"`
	Alias       string `json:"alias"`
}

func (v connectorVersion) columns() []string {
	return []string{v.ConnectorID, strconv.Itoa(v.Version)}
}

var connectorVersionColumns = []string{"Connector ID", "Version"}

// tag is an anchor point pointing to a version of the connector
type tag struct {
	ID            string `json:"id"`
	TagName       string `json:"tagName"`
	ActiveVersion uint32 `json:"activeVersion"`
}

func (t tag) columns() []string {
	return []string{t.ID, t.TagName, fmt.Sprint(t.ActiveVersion)}
}

var tagColumns = []string{"ID", "Tag Name", "Active Version"}

type TagCreate struct {
	TagName       string `json:"tagName"`
	ActiveVersion uint32 `json:"activeVersion"`
}

type TagUpdate struct {
	ActiveVersion uint32 `json:"activeVersion"`
}
