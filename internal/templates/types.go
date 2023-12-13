package templates

import (
	sailpointbetasdk "github.com/sailpoint-oss/golang-sdk/v2/beta"
	sailpointsdk "github.com/sailpoint-oss/golang-sdk/v2/v3"
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
	return len(template.Variables)
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
	return len(template.Variables)
}

type ReportQuery struct {
	QueryString string `json:"queryString"`
	QueryTitle  string `json:"queryTitle"`
	ResultCount string
}

type ReportTemplate struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Variables   []Variable    `json:"variables"`
	Queries     []ReportQuery `json:"queries"`
	Raw         []byte
}

func (template ReportTemplate) GetName() string {
	return template.Name
}

func (template ReportTemplate) GetDescription() string {
	return template.Description
}

func (template ReportTemplate) GetVariableCount() int {
	return len(template.Variables)
}

type Templates struct {
	Templates []SearchTemplate `json:"templates"`
}
