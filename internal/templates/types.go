package templates

import (
	sailpointbetasdk "github.com/sailpoint-oss/golang-sdk/beta"
	sailpointsdk "github.com/sailpoint-oss/golang-sdk/v3"
)

type Template interface {
	GetName() string
	GetDescription() string
	GetVariableCount() int
}

type Variable struct {
	Name   string `json:"name"`
	Prompt string `json:"prompt"`
}

type SearchTemplate struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Variables   []Variable          `json:"variables"`
	SearchQuery sailpointsdk.Search `json:"searchQuery"`
	Raw         []byte
}

func (template SearchTemplate) GetName() string {
	return template.Name
}

func (template SearchTemplate) GetDescription() string {
	return template.Description
}

func (template SearchTemplate) GetVariableCount() int {
	return len(template.Description)
}

type ExportTemplate struct {
	Name        string                         `json:"name"`
	Description string                         `json:"description"`
	Variables   []Variable                     `json:"variables"`
	ExportBody  sailpointbetasdk.ExportPayload `json:"exportBody"`
	Raw         []byte
}

func (template ExportTemplate) GetName() string {
	return template.Name
}

func (template ExportTemplate) GetDescription() string {
	return template.Description
}

func (template ExportTemplate) GetVariableCount() int {
	return len(template.Description)
}

type Templates struct {
	Templates []SearchTemplate `json:"templates"`
}
