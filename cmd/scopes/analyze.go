package scopes

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sailpoint-oss/sailpoint-cli/cmd/ui_plugins"
	"github.com/sailpoint-oss/sailpoint-cli/internal/output"
	scopeanalysis "github.com/sailpoint-oss/sailpoint-cli/internal/scopes"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

//go:embed analyze.md
var analyzeHelp string

type analyzeOptions struct {
	path           string
	authType       string
	jsonOutput     bool
	scopesOnly     bool
	updateManifest string
	// manifests overrides the scope data in tests.
	manifests map[scopeanalysis.Language]*scopeanalysis.Manifest
	locate    func(scopeanalysis.Language, []string) (string, bool)
}

func newAnalyzeCommand() *cobra.Command {
	opts := analyzeOptions{}

	help := util.ParseHelp(analyzeHelp)
	cmd := &cobra.Command{
		Use:     "analyze",
		Short:   "Find the least privilege scopes for the SailPoint SDK calls in a project",
		Long:    help.Long,
		Example: help.Example,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAnalyze(opts, cmd.OutOrStdout(), cmd.ErrOrStderr())
		},
	}

	cmd.Flags().StringVarP(&opts.path, "path", "p", ".", "Directory to scan")
	cmd.Flags().StringVar(&opts.authType, "auth-type", string(scopeanalysis.AuthUser), "Token type to report scopes for: user or application")
	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "Print the full report as JSON")
	cmd.Flags().BoolVar(&opts.scopesOnly, "scopes-only", false, "Print only the required scopes as a comma-separated list")
	cmd.Flags().StringVar(&opts.updateManifest, "update-manifest", "", "Write the required scopes into apiScopes of a UI plugin manifest (default: "+ui_plugins.ManifestFileName+" in --path)")
	cmd.Flags().Lookup("update-manifest").NoOptDefVal = ui_plugins.ManifestFileName
	cmd.MarkFlagsMutuallyExclusive("json", "scopes-only")

	return cmd
}

func runAnalyze(opts analyzeOptions, out, errOut io.Writer) error {
	auth, err := scopeanalysis.ParseAuthType(opts.authType)
	if err != nil {
		return err
	}

	res, err := scopeanalysis.Analyze(scopeanalysis.Options{
		Root:      opts.path,
		Auth:      auth,
		Manifests: opts.manifests,
		Locate:    opts.locate,
	})
	if err != nil {
		return err
	}

	switch {
	case opts.jsonOutput:
		encoded, err := json.MarshalIndent(res, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to encode report: %w", err)
		}
		_, _ = fmt.Fprintln(out, string(encoded))
	case opts.scopesOnly:
		_, _ = fmt.Fprintln(out, strings.Join(res.RequiredScopes, ","))
		for _, w := range res.Warnings {
			_, _ = fmt.Fprintf(errOut, "Warning: %s\n", w)
		}
	default:
		renderReport(out, res)
	}

	if opts.updateManifest != "" {
		return updateManifest(opts, res, out, errOut)
	}
	return nil
}

func renderReport(out io.Writer, res *scopeanalysis.Result) {
	if len(res.Calls) == 0 {
		_, _ = fmt.Fprintln(out, "No SailPoint SDK calls found.")
		return
	}

	byID := map[string]scopeanalysis.OperationResult{}
	for _, op := range res.Operations {
		byID[op.OperationID] = op
	}
	rows := make([][]string, 0, len(res.Calls))
	for _, c := range res.Calls {
		rows = append(rows, []string{
			c.File + ":" + strconv.Itoa(c.Line),
			string(c.Language),
			c.OperationID,
			describeScopes(byID[c.OperationID]),
		})
	}
	output.WriteTable(out, []string{"Location", "Language", "Operation", "Scopes"}, rows, "")

	_, _ = fmt.Fprintf(out, "\nRequired scopes (%s auth):\n", res.AuthType)
	if len(res.RequiredScopes) == 0 {
		_, _ = fmt.Fprintln(out, "  (none)")
	}
	for _, s := range res.RequiredScopes {
		_, _ = fmt.Fprintf(out, "  %s\n", s)
	}

	_, _ = fmt.Fprintln(out, "\nScope data:")
	for _, src := range res.Sources {
		_, _ = fmt.Fprintf(out, "  %s: %s\n", src.Language, describeSource(src))
	}

	if len(res.Warnings) > 0 {
		_, _ = fmt.Fprintln(out, "\nWarnings:")
		for _, w := range res.Warnings {
			_, _ = fmt.Fprintf(out, "  %s\n", w)
		}
	}
}

func describeScopes(op scopeanalysis.OperationResult) string {
	switch op.Status {
	case scopeanalysis.StatusUnsupported:
		return "(not available for this auth type)"
	case scopeanalysis.StatusNoScope:
		return "(no scope needed)"
	}
	return strings.Join(op.Alternatives, " | ")
}

func describeSource(src scopeanalysis.Source) string {
	switch {
	case src.Embedded:
		return "embedded in the CLI (" + src.Detail + ")"
	case src.Path != "" && src.SDKVersion != "":
		return fmt.Sprintf("SDK %s (%s)", src.SDKVersion, src.Path)
	case src.Path != "":
		return src.Path
	}
	return src.Detail
}

func updateManifest(opts analyzeOptions, res *scopeanalysis.Result, out, errOut io.Writer) error {
	path := opts.updateManifest
	if path == ui_plugins.ManifestFileName {
		path = filepath.Join(opts.path, path)
	}
	if len(res.RequiredScopes) == 0 {
		_, _ = fmt.Fprintf(errOut, "No scopes found; %s was not updated.\n", path)
		return nil
	}
	if err := ui_plugins.UpdateManifestAPIScopes(path, res.RequiredScopes); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(errOut, "Updated apiScopes in %s.\n", path)
	return nil
}
