// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package transform

type transform struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

func (t transform) transformToColumns() []string {
	return []string{t.ID, t.Name}
}

var transformColumns = []string{"ID", "Name"}
