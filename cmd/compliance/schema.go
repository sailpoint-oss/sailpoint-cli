package compliance

import (
	"encoding/json"
	"time"
)

type EvidenceBundle struct {
	Metadata Metadata          `json:"metadata"`
	Data     EvidenceData      `json:"data"`
	Summary  CollectionSummary `json:"summary"`
}

type Metadata struct {
	SchemaVersion  string    `json:"schema_version"`
	GeneratedAt    time.Time `json:"generated_at"`
	SailCLIVersion string    `json:"sail_cli_version"`
	PeriodDays     int       `json:"period_days"`
	Tenant         string    `json:"tenant"`
}

type EvidenceData struct {
	AuthOrgConfig    json.RawMessage `json:"auth_org_config,omitempty"`
	PasswordPolicies json.RawMessage `json:"password_policies,omitempty"`
	SODPolicies      json.RawMessage `json:"sod_policies,omitempty"`
	Certifications   json.RawMessage `json:"certifications,omitempty"`
	Identities       json.RawMessage `json:"identities,omitempty"`
	Roles            json.RawMessage `json:"roles,omitempty"`
	AccessProfiles   json.RawMessage `json:"access_profiles,omitempty"`
	Sources          json.RawMessage `json:"sources,omitempty"`
	LifecycleStates  json.RawMessage `json:"lifecycle_states,omitempty"`
	Workflows        json.RawMessage `json:"workflows,omitempty"`
	GovernanceGroups json.RawMessage `json:"governance_groups,omitempty"`
	Events           EventSummary    `json:"events"`
}

type EventSummary struct {
	ProvisioningCount int `json:"provisioning_count"`
	PasswordCount     int `json:"password_count"`
	PeriodDays        int `json:"period_days"`
}

type CollectionSummary struct {
	TotalCollectors int      `json:"total_collectors"`
	Succeeded       int      `json:"succeeded"`
	Failed          int      `json:"failed"`
	Errors          []string `json:"errors,omitempty"`
}

type EvaluationResult struct {
	Metadata Metadata        `json:"metadata"`
	Controls []ControlResult `json:"controls"`
	Findings []Finding       `json:"findings"`
	Summary  EvalSummary     `json:"summary"`
}

type ControlResult struct {
	ControlID    string        `json:"control_id"`
	ControlTitle string        `json:"control_title"`
	Status       string        `json:"status"`
	Checks       []CheckResult `json:"checks"`
}

type CheckResult struct {
	CheckID     string      `json:"check_id"`
	Description string      `json:"description"`
	Status      string      `json:"status"`
	Severity    string      `json:"severity"`
	Expected    interface{} `json:"expected,omitempty"`
	Actual      interface{} `json:"actual,omitempty"`
	Remediation string      `json:"remediation,omitempty"`
}

type Finding struct {
	ControlID   string `json:"control_id"`
	CheckID     string `json:"check_id"`
	Severity    string `json:"severity"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type EvalSummary struct {
	TotalControls    int `json:"total_controls"`
	Passed           int `json:"passed"`
	Failed           int `json:"failed"`
	CriticalFindings int `json:"critical_findings"`
	HighFindings     int `json:"high_findings"`
}

type ControlPack struct {
	Controls []ControlDefinition `yaml:"controls"`
}

type ControlDefinition struct {
	ControlID    string            `yaml:"control_id"`
	ControlTitle string            `yaml:"control_title"`
	Checks       []CheckDefinition `yaml:"checks"`
}

type CheckDefinition struct {
	CheckID     string      `yaml:"check_id"`
	Description string      `yaml:"description"`
	Severity    string      `yaml:"severity"`
	JSONPath    string      `yaml:"json_path"`
	Rule        string      `yaml:"rule"`
	Field       string      `yaml:"field,omitempty"`
	Expected    interface{} `yaml:"expected,omitempty"`
	Pattern     string      `yaml:"pattern,omitempty"`
	Remediation string      `yaml:"remediation,omitempty"`
}
