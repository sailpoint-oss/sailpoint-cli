package validate

import (
	"context"
	"fmt"

	"github.com/sailpoint/sp-cli/client"
)

var Checks = []Check{}

func init() {
	Checks = append(Checks, accountCreateChecks...)
	Checks = append(Checks, accountReadChecks...)
	Checks = append(Checks, accountUpdateChecks...)
	Checks = append(Checks, entitlementReadChecks...)
	Checks = append(Checks, testConnChecks...)
}

// Check represents a specific property we want to validate
type Check struct {
	ID          string
	Description string

	// IsDataModifier determines a checking that will modify connectors data after applying
	IsDataModifier bool
	Run            func(ctx context.Context, spec *client.ConnSpec, cc *client.ConnClient, res *CheckResult)
	// RequiredCommands represents a list of commands that use for this check
	RequiredCommands []string
}

// CheckResult captures the result of an individual check.
type CheckResult struct {
	// ID is a short human readable slug describing the check
	ID string

	// Errors is a list of errors encountered when running the test.
	Errors []string

	// Skipped is a short description why the check was skipped
	Skipped []string

	// Warnings is a list of warnings encountered when running the test.
	Warnings []string
}

// err adds the provided err to the list of errors for the check
func (res *CheckResult) err(err error) {
	res.Errors = append(res.Errors, err.Error())
}

// errf adds an error to the check result
func (res *CheckResult) errf(format string, a ...interface{}) {
	res.Errors = append(res.Errors, fmt.Sprintf(format, a...))
}

// warnf adds an warning to the check result
func (res *CheckResult) warnf(format string, a ...interface{}) {
	res.Warnings = append(res.Warnings, fmt.Sprintf(format, a...))
}

// skipf adds a reason of a skipped check to the check result
func (res *CheckResult) skipf(format string, a ...interface{}) {
	res.Skipped = append(res.Skipped, fmt.Sprintf(format, a...))
}
