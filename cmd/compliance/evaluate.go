package compliance

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

//go:embed evaluate.md
var evaluateHelp string

func newEvaluateCommand() *cobra.Command {
	help := util.ParseHelp(evaluateHelp)

	var inputFile string
	var controls string
	var outputFile string
	var outputMarkdown string

	cmd := &cobra.Command{
		Use:     "evaluate",
		Short:   "Evaluate an evidence bundle against a control pack",
		Long:    help.Long,
		Example: help.Example,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			evidenceRaw, err := os.ReadFile(inputFile)
			if err != nil {
				return fmt.Errorf("failed to read evidence file %q: %w", inputFile, err)
			}

			var evidence EvidenceBundle
			if err := json.Unmarshal(evidenceRaw, &evidence); err != nil {
				return fmt.Errorf("failed to parse evidence bundle: %w", err)
			}

			pack, resolvedControls, err := loadControlPack(controls)
			if err != nil {
				return err
			}

			result := evaluateControlPack(evidenceRaw, evidence, pack)
			if result.Metadata.GeneratedAt.IsZero() {
				result.Metadata.GeneratedAt = time.Now().UTC()
			}

			if err := writeJSONOutput(outputFile, result, true); err != nil {
				return err
			}

			if strings.TrimSpace(outputMarkdown) != "" {
				report := buildMarkdownReport(result, resolvedControls)
				if err := writeTextOutput(outputMarkdown, report); err != nil {
					return err
				}
			}

			if result.Summary.Failed > 0 || len(result.Findings) > 0 {
				return fmt.Errorf("evaluation failed: %d controls failed/partial, %d findings", result.Summary.Failed, len(result.Findings))
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&inputFile, "input", "i", "", "Path to input evidence JSON")
	cmd.Flags().StringVarP(&controls, "controls", "c", "nist-800-53", "Control pack name or path to control YAML")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "findings.json", "Path to output findings JSON")
	cmd.Flags().StringVar(&outputMarkdown, "output-md", "", "Optional path to output markdown report")
	cmd.MarkFlagRequired("input")

	return cmd
}

func buildMarkdownReport(result EvaluationResult, controlPack string) string {
	var b strings.Builder

	b.WriteString("# Compliance Evaluation Report\n\n")
	b.WriteString(fmt.Sprintf("- Generated At: `%s`\n", time.Now().UTC().Format(time.RFC3339)))
	b.WriteString(fmt.Sprintf("- Tenant: `%s`\n", result.Metadata.Tenant))
	b.WriteString(fmt.Sprintf("- Controls: `%s`\n", controlPack))
	b.WriteString(fmt.Sprintf("- Evidence Generated At: `%s`\n\n", result.Metadata.GeneratedAt.Format(time.RFC3339)))

	b.WriteString("## Summary\n\n")
	b.WriteString(fmt.Sprintf("- Total controls: %d\n", result.Summary.TotalControls))
	b.WriteString(fmt.Sprintf("- Passed controls: %d\n", result.Summary.Passed))
	b.WriteString(fmt.Sprintf("- Failed/PARTIAL controls: %d\n", result.Summary.Failed))
	b.WriteString(fmt.Sprintf("- Critical findings: %d\n", result.Summary.CriticalFindings))
	b.WriteString(fmt.Sprintf("- High findings: %d\n\n", result.Summary.HighFindings))

	b.WriteString("## Controls\n\n")
	b.WriteString("| Control ID | Title | Status |\n")
	b.WriteString("| --- | --- | --- |\n")
	for _, control := range result.Controls {
		b.WriteString(fmt.Sprintf("| %s | %s | %s |\n", control.ControlID, control.ControlTitle, control.Status))
	}
	b.WriteString("\n")

	b.WriteString("## Findings\n\n")
	if len(result.Findings) == 0 {
		b.WriteString("No findings.\n")
		return b.String()
	}

	severityOrder := []string{"critical", "high", "medium", "low", "info"}
	for _, severity := range severityOrder {
		b.WriteString(fmt.Sprintf("### %s\n\n", strings.Title(severity)))
		count := 0
		for _, finding := range result.Findings {
			if normalizeSeverity(finding.Severity) != severity {
				continue
			}
			count++
			b.WriteString(fmt.Sprintf("- `%s/%s` %s\n", finding.ControlID, finding.CheckID, finding.Description))
		}
		if count == 0 {
			b.WriteString("- None\n")
		}
		b.WriteString("\n")
	}

	return b.String()
}

func writeTextOutput(outputPath string, content string) error {
	dir := filepath.Dir(outputPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(outputPath, []byte(content), 0o644)
}
