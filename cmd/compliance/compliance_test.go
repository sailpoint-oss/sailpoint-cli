package compliance

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewComplianceCommand(t *testing.T) {
	cmd := NewComplianceCommand()

	subcommands := map[string]bool{}
	for _, subcommand := range cmd.Commands() {
		subcommands[subcommand.Name()] = true
	}

	if !subcommands["collect"] {
		t.Fatalf("expected collect subcommand")
	}
	if !subcommands["evaluate"] {
		t.Fatalf("expected evaluate subcommand")
	}

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected command to execute without args: %v", err)
	}
}

func TestEvaluateWithEmbeddedControls(t *testing.T) {
	tmp := t.TempDir()
	inputPath := filepath.Join(tmp, "evidence.json")
	outputPath := filepath.Join(tmp, "findings.json")

	bundle := EvidenceBundle{
		Metadata: Metadata{
			SchemaVersion:  "1.0.0",
			GeneratedAt:    time.Now().UTC(),
			SailCLIVersion: "test",
			PeriodDays:     90,
			Tenant:         "https://example.identitynow.com",
		},
		Data: EvidenceData{
			AuthOrgConfig:    json.RawMessage(`{"lockoutThreshold": 5, "mfaEnabled": true}`),
			PasswordPolicies: json.RawMessage(`[{"minLength": 14}]`),
			SODPolicies:      json.RawMessage(`[{"id":"sod-1"}]`),
			Certifications:   json.RawMessage(`[{"id":"cert-1"}]`),
			Sources:          json.RawMessage(`[{"owner":{"id":"owner-1"},"authoritative":true}]`),
			LifecycleStates:  json.RawMessage(`[{"technicalName":"terminated"}]`),
			Events: EventSummary{
				ProvisioningCount: 2,
				PasswordCount:     1,
				PeriodDays:        90,
			},
		},
		Summary: CollectionSummary{
			TotalCollectors: 13,
			Succeeded:       13,
			Failed:          0,
		},
	}

	data, err := json.Marshal(bundle)
	if err != nil {
		t.Fatalf("failed to marshal evidence bundle: %v", err)
	}
	if err := os.WriteFile(inputPath, data, 0o644); err != nil {
		t.Fatalf("failed to write evidence fixture: %v", err)
	}

	cmd := newEvaluateCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--input", inputPath, "--output", outputPath})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected evaluation to pass, got error: %v\noutput: %s", err, buf.String())
	}

	resultData, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	var result EvaluationResult
	if err := json.Unmarshal(resultData, &result); err != nil {
		t.Fatalf("failed to parse evaluation output: %v", err)
	}

	if len(result.Controls) == 0 {
		t.Fatalf("expected at least one control result")
	}
	if result.Summary.Failed != 0 {
		t.Fatalf("expected no failed controls, got %d", result.Summary.Failed)
	}
	if len(result.Findings) != 0 {
		t.Fatalf("expected no findings, got %d", len(result.Findings))
	}
}

func TestEvaluateMalformedInput(t *testing.T) {
	tmp := t.TempDir()
	inputPath := filepath.Join(tmp, "bad-evidence.json")
	outputPath := filepath.Join(tmp, "findings.json")

	if err := os.WriteFile(inputPath, []byte("{not-valid-json"), 0o644); err != nil {
		t.Fatalf("failed to write malformed fixture: %v", err)
	}

	cmd := newEvaluateCommand()
	cmd.SetArgs([]string{"--input", inputPath, "--output", outputPath})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected malformed input to return error")
	}
	if !strings.Contains(err.Error(), "failed to parse evidence bundle") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestEvaluateMissingInputFlag(t *testing.T) {
	cmd := newEvaluateCommand()
	cmd.SetArgs([]string{"--output", "findings.json"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected missing input flag to return error")
	}
	if !strings.Contains(err.Error(), "required flag(s) \"input\" not set") {
		t.Fatalf("unexpected error for missing input flag: %v", err)
	}
}
