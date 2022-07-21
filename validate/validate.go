// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package validate

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/sailpoint/sp-cli/client"
)

// Validator runs checks for a specific connector
type Validator struct {
	cfg      Config
	cc       *client.ConnClient
	connSpec *client.ConnSpec
}

// Config provides options for how the validator runs
type Config struct {
	// Check specifies a single check that should be run. If this is empty then
	// all checks are run.
	Check string

	// ReadOnly specifies a type of validation.
	// If ReadOnly set 'true' validator will run all checks that don't make any modifications.
	ReadOnly bool
}

// NewValidator creates a new validator with provided config and ConnClient
func NewValidator(cfg Config, cc *client.ConnClient) *Validator {
	return &Validator{
		cfg: cfg,
		cc:  cc,
	}
}

// Run runs the validator suite
func (v *Validator) Run(ctx context.Context) (results []CheckResult, err error) {
	rand.Seed(time.Now().UnixNano())

	spec, err := v.cc.SpecRead(ctx)
	if err != nil {
		return nil, err
	}
	for _, check := range Checks {
		if v.cfg.ReadOnly && check.IsDataModifier {
			continue
		}

		if len(v.cfg.Check) > 0 && check.ID != v.cfg.Check {
			continue
		}

		log.Printf("running check %q", check.ID)

		res := &CheckResult{
			ID: check.ID,
		}

		if ok, results := isCheckPossible(spec.Commands, check.RequiredCommands); ok {
			check.Run(ctx, spec, v.cc, res)
		} else {
			res.skipf("Skipping check due to unimplemented commands on a connector: %s", strings.Join(results, ", "))
		}

		results = append(results, *res)
		fmt.Println()
	}
	return results, nil
}
