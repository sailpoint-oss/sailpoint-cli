package compliance

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/sailpoint-oss/sailpoint-cli/internal/jsonpath"
	"gopkg.in/yaml.v2"
)

//go:embed controls/nist_800_53.yaml
var defaultControlPackYAML []byte

func loadControlPack(controlsArg string) (ControlPack, string, error) {
	var payload []byte
	resolved := strings.TrimSpace(controlsArg)
	if resolved == "" {
		resolved = "nist-800-53"
	}

	if isDefaultControlPack(resolved) {
		payload = defaultControlPackYAML
		resolved = "nist-800-53"
	} else {
		custom, err := osReadFile(resolved)
		if err != nil {
			return ControlPack{}, "", fmt.Errorf("failed to read controls file %q: %w", resolved, err)
		}
		payload = custom
	}

	var pack ControlPack
	if err := yaml.Unmarshal(payload, &pack); err != nil {
		return ControlPack{}, "", fmt.Errorf("failed to parse control pack: %w", err)
	}

	if err := validateControlPack(pack); err != nil {
		return ControlPack{}, "", err
	}

	return pack, resolved, nil
}

func evaluateControlPack(evidenceRaw []byte, evidence EvidenceBundle, pack ControlPack) EvaluationResult {
	result := EvaluationResult{
		Metadata: evidence.Metadata,
		Controls: make([]ControlResult, 0, len(pack.Controls)),
		Findings: []Finding{},
	}

	for _, control := range pack.Controls {
		controlResult := ControlResult{
			ControlID:    control.ControlID,
			ControlTitle: control.ControlTitle,
			Checks:       make([]CheckResult, 0, len(control.Checks)),
		}

		for _, check := range control.Checks {
			checkResult := evaluateCheck(evidenceRaw, check)
			controlResult.Checks = append(controlResult.Checks, checkResult)

			if checkResult.Status == "FAIL" {
				result.Findings = append(result.Findings, Finding{
					ControlID:   control.ControlID,
					CheckID:     check.CheckID,
					Severity:    normalizeSeverity(check.Severity),
					Title:       fmt.Sprintf("%s %s failed", control.ControlID, check.CheckID),
					Description: check.Description,
				})
			}
		}

		controlResult.Status = deriveControlStatus(controlResult.Checks)
		result.Controls = append(result.Controls, controlResult)
	}

	result.Summary = summarizeEvaluation(result)
	return result
}

func evaluateCheck(evidenceRaw []byte, check CheckDefinition) CheckResult {
	severity := normalizeSeverity(check.Severity)
	result := CheckResult{
		CheckID:     check.CheckID,
		Description: check.Description,
		Severity:    severity,
		Expected:    check.Expected,
		Remediation: check.Remediation,
	}

	targetRaw, err := jsonpath.EvaluateJSONPath(evidenceRaw, check.JSONPath)
	if err != nil {
		result.Status = "NOT_ASSESSED"
		result.Actual = err.Error()
		return result
	}

	var target interface{}
	if err := json.Unmarshal(targetRaw, &target); err != nil {
		result.Status = "NOT_ASSESSED"
		result.Actual = fmt.Sprintf("invalid JSONPath result: %v", err)
		return result
	}

	pass, actual, err := runRule(target, check)
	if actual != nil {
		result.Actual = actual
	}
	if err != nil {
		result.Status = "NOT_ASSESSED"
		if result.Actual == nil {
			result.Actual = err.Error()
		}
		return result
	}

	if pass {
		result.Status = "PASS"
	} else {
		result.Status = "FAIL"
	}

	return result
}

func runRule(target interface{}, check CheckDefinition) (bool, interface{}, error) {
	switch strings.ToLower(strings.TrimSpace(check.Rule)) {
	case "all_have_field":
		items, ok := asSlice(target)
		if !ok {
			return false, nil, fmt.Errorf("rule all_have_field requires array target")
		}
		matching := 0
		for _, item := range items {
			if _, ok := getFieldValue(item, check.Field); ok {
				matching++
			}
		}
		actual := map[string]interface{}{"total": len(items), "matching": matching}
		return len(items) > 0 && matching == len(items), actual, nil

	case "any_match":
		items, ok := asSlice(target)
		if !ok {
			return false, nil, fmt.Errorf("rule any_match requires array target")
		}
		matching := 0
		for _, item := range items {
			value, exists := getFieldValue(item, check.Field)
			if exists && valuesEqual(value, check.Expected) {
				matching++
			}
		}
		actual := map[string]interface{}{"total": len(items), "matching": matching}
		return matching > 0, actual, nil

	case "count_gte":
		expected, ok := toFloat64(check.Expected)
		if !ok {
			return false, nil, fmt.Errorf("rule count_gte requires numeric expected")
		}
		count, ok := collectionCount(target)
		if !ok {
			return false, nil, fmt.Errorf("rule count_gte requires array/object target")
		}
		return float64(count) >= expected, count, nil

	case "field_exists":
		value, ok := getFieldValue(target, check.Field)
		if ok {
			return true, value, nil
		}
		return false, nil, nil

	case "field_equals":
		value, ok := getFieldValue(target, check.Field)
		if !ok {
			return false, nil, nil
		}
		return valuesEqual(value, check.Expected), value, nil

	case "all_field_gte":
		expected, ok := toFloat64(check.Expected)
		if !ok {
			return false, nil, fmt.Errorf("rule all_field_gte requires numeric expected")
		}
		items, ok := asSlice(target)
		if !ok {
			return false, nil, fmt.Errorf("rule all_field_gte requires array target")
		}
		matching := 0
		for _, item := range items {
			value, exists := getFieldValue(item, check.Field)
			if !exists {
				continue
			}
			numeric, isNumeric := toFloat64(value)
			if !isNumeric {
				return false, nil, fmt.Errorf("field %q must be numeric for all_field_gte", check.Field)
			}
			if numeric >= expected {
				matching++
			}
		}
		actual := map[string]interface{}{"total": len(items), "matching": matching}
		return len(items) > 0 && matching == len(items), actual, nil

	case "any_field_matches":
		if check.Pattern == "" {
			return false, nil, fmt.Errorf("rule any_field_matches requires pattern")
		}
		re, err := regexp.Compile(check.Pattern)
		if err != nil {
			return false, nil, fmt.Errorf("invalid regex pattern: %w", err)
		}
		items, ok := asSlice(target)
		if !ok {
			return false, nil, fmt.Errorf("rule any_field_matches requires array target")
		}
		matching := 0
		for _, item := range items {
			value, exists := getFieldValue(item, check.Field)
			if !exists {
				continue
			}
			if re.MatchString(fmt.Sprintf("%v", value)) {
				matching++
			}
		}
		actual := map[string]interface{}{"total": len(items), "matching": matching}
		return matching > 0, actual, nil

	case "value_gte":
		expected, ok := toFloat64(check.Expected)
		if !ok {
			return false, nil, fmt.Errorf("rule value_gte requires numeric expected")
		}
		actual, ok := toFloat64(target)
		if !ok {
			return false, nil, fmt.Errorf("rule value_gte requires numeric target")
		}
		return actual >= expected, actual, nil

	default:
		return false, nil, fmt.Errorf("unsupported rule %q", check.Rule)
	}
}

func validateControlPack(pack ControlPack) error {
	if len(pack.Controls) == 0 {
		return fmt.Errorf("control pack has no controls")
	}

	for _, control := range pack.Controls {
		if strings.TrimSpace(control.ControlID) == "" {
			return fmt.Errorf("control is missing control_id")
		}
		if strings.TrimSpace(control.ControlTitle) == "" {
			return fmt.Errorf("control %s is missing control_title", control.ControlID)
		}
		if len(control.Checks) == 0 {
			return fmt.Errorf("control %s has no checks", control.ControlID)
		}
		for _, check := range control.Checks {
			if err := validateCheck(check, control.ControlID); err != nil {
				return err
			}
		}
	}

	return nil
}

func validateCheck(check CheckDefinition, controlID string) error {
	if strings.TrimSpace(check.CheckID) == "" {
		return fmt.Errorf("control %s has check with missing check_id", controlID)
	}
	if strings.TrimSpace(check.Description) == "" {
		return fmt.Errorf("control %s check %s is missing description", controlID, check.CheckID)
	}
	if strings.TrimSpace(check.Severity) == "" {
		return fmt.Errorf("control %s check %s is missing severity", controlID, check.CheckID)
	}
	if strings.TrimSpace(check.JSONPath) == "" {
		return fmt.Errorf("control %s check %s is missing json_path", controlID, check.CheckID)
	}
	if strings.TrimSpace(check.Rule) == "" {
		return fmt.Errorf("control %s check %s is missing rule", controlID, check.CheckID)
	}

	rule := strings.ToLower(strings.TrimSpace(check.Rule))
	requiresField := map[string]bool{
		"all_have_field":    true,
		"any_match":         true,
		"field_exists":      true,
		"field_equals":      true,
		"all_field_gte":     true,
		"any_field_matches": true,
	}
	if requiresField[rule] && strings.TrimSpace(check.Field) == "" {
		return fmt.Errorf("control %s check %s rule %s requires field", controlID, check.CheckID, rule)
	}

	requiresExpected := map[string]bool{
		"any_match":     true,
		"count_gte":     true,
		"field_equals":  true,
		"all_field_gte": true,
		"value_gte":     true,
	}
	if requiresExpected[rule] && check.Expected == nil {
		return fmt.Errorf("control %s check %s rule %s requires expected", controlID, check.CheckID, rule)
	}

	if rule == "any_field_matches" && strings.TrimSpace(check.Pattern) == "" {
		return fmt.Errorf("control %s check %s rule any_field_matches requires pattern", controlID, check.CheckID)
	}

	return nil
}

func deriveControlStatus(checks []CheckResult) string {
	if len(checks) == 0 {
		return "NOT_ASSESSED"
	}

	pass := 0
	fail := 0
	notAssessed := 0

	for _, check := range checks {
		switch check.Status {
		case "PASS":
			pass++
		case "FAIL":
			fail++
		default:
			notAssessed++
		}
	}

	if pass == len(checks) {
		return "PASS"
	}
	if notAssessed == len(checks) {
		return "NOT_ASSESSED"
	}
	if pass == 0 && fail > 0 {
		return "FAIL"
	}
	return "PARTIAL"
}

func summarizeEvaluation(result EvaluationResult) EvalSummary {
	summary := EvalSummary{TotalControls: len(result.Controls)}

	for _, control := range result.Controls {
		switch control.Status {
		case "PASS":
			summary.Passed++
		case "FAIL", "PARTIAL":
			summary.Failed++
		}
	}

	for _, finding := range result.Findings {
		switch normalizeSeverity(finding.Severity) {
		case "critical":
			summary.CriticalFindings++
		case "high":
			summary.HighFindings++
		}
	}

	return summary
}

func asSlice(value interface{}) ([]interface{}, bool) {
	items, ok := value.([]interface{})
	return items, ok
}

func collectionCount(value interface{}) (int, bool) {
	switch typed := value.(type) {
	case []interface{}:
		return len(typed), true
	case map[string]interface{}:
		return len(typed), true
	default:
		return 0, false
	}
}

func getFieldValue(value interface{}, fieldPath string) (interface{}, bool) {
	if strings.TrimSpace(fieldPath) == "" {
		return nil, false
	}

	parts := strings.Split(fieldPath, ".")
	current := value
	for _, part := range parts {
		obj, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}
		next, exists := obj[part]
		if !exists || next == nil {
			return nil, false
		}
		current = next
	}

	if str, ok := current.(string); ok && strings.TrimSpace(str) == "" {
		return nil, false
	}

	return current, true
}

func toFloat64(value interface{}) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int8:
		return float64(typed), true
	case int16:
		return float64(typed), true
	case int32:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case uint:
		return float64(typed), true
	case uint8:
		return float64(typed), true
	case uint16:
		return float64(typed), true
	case uint32:
		return float64(typed), true
	case uint64:
		return float64(typed), true
	case json.Number:
		parsed, err := typed.Float64()
		return parsed, err == nil
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		return parsed, err == nil
	default:
		return 0, false
	}
}

func valuesEqual(actual interface{}, expected interface{}) bool {
	if actualNum, ok := toFloat64(actual); ok {
		if expectedNum, ok := toFloat64(expected); ok {
			return actualNum == expectedNum
		}
	}

	return fmt.Sprintf("%v", actual) == fmt.Sprintf("%v", expected)
}

func normalizeSeverity(severity string) string {
	normalized := strings.ToLower(strings.TrimSpace(severity))
	switch normalized {
	case "critical", "high", "medium", "low", "info":
		return normalized
	default:
		return "medium"
	}
}

func isDefaultControlPack(input string) bool {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "nist-800-53", "nist_800_53", "nist80053", "nist-800-53-r5", "default":
		return true
	default:
		return false
	}
}

// osReadFile is wrapped for easy stubbing in tests.
var osReadFile = func(path string) ([]byte, error) {
	return os.ReadFile(path)
}
