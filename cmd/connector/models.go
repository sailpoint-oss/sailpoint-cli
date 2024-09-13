// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package connector

import (
	"fmt"
	"strconv"
)

type connector struct {
	ID    string `json:"id"`
	Alias string `json:"alias"`
}

func (c connector) columns() []string {
	return []string{c.ID, c.Alias}
}

var connectorColumns = []string{"ID", "Alias"}

type connectorList struct {
	ID            string `json:"id"`
	Alias         string `json:"alias"`
	TagName       string `json:"tagName"`
	ActiveVersion uint32 `json:"activeVersion"`
}

func (c connectorList) columns() []string {
	return []string{c.ID, c.Alias, c.TagName, fmt.Sprint(c.ActiveVersion)}
}

var connectorListColumns = []string{"ID", "Alias", "Tags", "Version"}

type connectorVersion struct {
	ConnectorID string `json:"connectorId"`
	Version     int    `json:"version"`
}

type connectorUpdate struct {
	Alias string `json:"alias"`
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

type instance struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	CustomizerId string `json:"connectorCustomizerId"`
}

func (c instance) columns() []string {
	return []string{c.ID, c.Name, c.CustomizerId}
}

var instanceColumns = []string{"ID", "Name", "Customizer ID"}

type customizer struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ImageVersion *int   `json:"imageVersion,omitempty"`
}

func (c customizer) columns() []string {
	if c.ImageVersion == nil {
		return []string{c.ID, c.Name, ""}
	}
	return []string{c.ID, c.Name, strconv.Itoa(*c.ImageVersion)}
}

var customizerColumns = []string{"ID", "Name", "Version"}

type customizerVersion struct {
	CustomizerID string `json:"connectorCustomizerId"`
	ImageID      string `json:"imageId"`
	Version      int    `json:"version"`
}

func (c customizerVersion) columns() []string {
	return []string{c.CustomizerID, c.ImageID, strconv.Itoa(c.Version)}
}

var customizerVersionColumns = []string{"Customizer ID", "Image ID", "Version"}
