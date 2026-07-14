package ui_plugins

import (
	"bufio"
	"context"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/sailpoint-oss/sailpoint-cli/internal/initialize"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

//go:embed init.md
var initHelp string

const (
	uiPluginTemplatesRepoOwner = "sailpoint-oss"
	uiPluginTemplatesRepoName  = "ui-plugin-templates"
	uiPluginTemplatesBranch    = "main"
	angularStarterSubdir       = "angular/starter"
	genericGuideFileName       = "SAILPOINT_PLUGIN_GUIDE.md"
	defaultDevServerPort       = 3000

	// maxAliasPrompts bounds the interactive re-prompt loop so exhausted stdin
	// (EOF) can't spin it forever.
	maxAliasPrompts = 5
)

// initOptions holds the resolved flag values for the init command.
type initOptions struct {
	name    string
	alias   string
	path    string
	outDir  string
	port    int
	portSet bool
	force   bool
}

// initDeps carries the command's I/O plus the network-touching seams (scaffold
// and fetchGuide) so the orchestration is testable without GitHub or a live
// tenant.
type initDeps struct {
	client     client.Client
	in         io.Reader
	out        io.Writer
	errOut     io.Writer
	scaffold   func(destDir string) error  // fetch + extract the angular starter subtree
	fetchGuide func(destPath string) error // fetch + write the generic guide
}

func newInitCommand() *cobra.Command {
	var opts initOptions

	help := util.ParseHelp(initHelp)
	cmd := &cobra.Command{
		Use:     "init [plugin-name]",
		Short:   "Scaffold a new UI plugin workspace or attach the SDK to an existing project",
		Long:    help.Long,
		Example: help.Example,
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
				opts.name = args[0]
			}
			opts.name = strings.TrimSpace(opts.name)
			opts.alias = strings.TrimSpace(opts.alias)
			opts.path = strings.TrimSpace(opts.path)
			opts.outDir = strings.TrimSpace(opts.outDir)
			opts.portSet = cmd.Flags().Changed("port")

			// Alias validation is best-effort: acquire an authenticated client if
			// we can, but continue (with a warning) when one is unavailable so a
			// misconfigured or offline client doesn't block getting to a working
			// local workspace. A definitive conflict/invalid answer still fails.
			spClient, err := newPluginClient()
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: could not initialize an authenticated client; alias will not be validated: %v\n", err)
				spClient = nil
			}

			deps := initDeps{
				client:     spClient,
				in:         cmd.InOrStdin(),
				out:        cmd.OutOrStdout(),
				errOut:     cmd.ErrOrStderr(),
				scaffold:   defaultScaffold,
				fetchGuide: defaultFetchGuide,
			}
			return runInit(context.Background(), deps, opts)
		},
	}

	cmd.Flags().StringVar(&opts.name, "name", "", "Plugin display name")
	cmd.Flags().StringVar(&opts.alias, "alias", "", "Tenant-unique plugin alias (defaults to a slug of the name)")
	cmd.Flags().StringVar(&opts.path, "path", "", "Attach the SDK to the existing project at this directory instead of scaffolding a new workspace")
	cmd.Flags().StringVar(&opts.outDir, "out-dir", "", "Build output directory (required when attaching with --path)")
	cmd.Flags().IntVar(&opts.port, "port", defaultDevServerPort, "Local dev server port (used when attaching with --path)")
	cmd.Flags().BoolVar(&opts.force, "force", false, "Overwrite plugin files this command manages if they already exist")

	return cmd
}

// runInit resolves inputs, validates the alias against UMS (before any
// filesystem mutation), then dispatches to the new-workspace or --path flow.
// The command runs headless (no prompts) when the required flags for the chosen
// flow are supplied; otherwise it prompts interactively.
func runInit(ctx context.Context, deps initDeps, opts initOptions) error {
	r := bufio.NewReader(deps.in)

	// A provided --name (or positional) is the signal to run headless; without
	// it we prompt for the missing inputs.
	interactive := opts.name == ""

	name, err := resolveName(r, deps, opts, interactive)
	if err != nil {
		return err
	}
	alias, err := resolveAndCheckAlias(ctx, r, deps, opts, name, interactive)
	if err != nil {
		return err
	}

	if opts.path != "" {
		return runPathAttach(r, deps, opts, name, alias, interactive)
	}
	return runNewWorkspace(deps, name, alias)
}

// runNewWorkspace scaffolds the Angular starter into ./<alias>/ and personalizes
// it. It hard-fails if the destination already exists.
func runNewWorkspace(deps initDeps, name, alias string) error {
	destDir := alias
	if _, err := os.Stat(destDir); err == nil {
		return fmt.Errorf("error: project '%s' already exists", destDir)
	}

	if err := deps.scaffold(destDir); err != nil {
		// Clean up a partially extracted directory; it did not exist before.
		_ = os.RemoveAll(destDir)
		return err
	}

	if err := personalizeScaffold(destDir, alias, name); err != nil {
		return err
	}

	fmt.Fprintf(deps.out, "\nCreated plugin workspace %q.\n\nNext steps:\n  cd %s\n  npm install\n  sail ui-plugins create\n", alias, destDir)
	return nil
}

// runPathAttach makes an existing project SDK-ready without touching unrelated
// files: it generates sp-ui-plugin.json and drops the generic guide, prompting
// (or honoring --force) before overwriting either file it manages.
func runPathAttach(r *bufio.Reader, deps initDeps, opts initOptions, name, alias string, interactive bool) error {
	info, err := os.Stat(opts.path)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("--path %q is not an existing directory", opts.path)
	}

	outDir, err := resolveOutDir(r, deps, opts, interactive)
	if err != nil {
		return err
	}
	port, err := resolvePort(r, deps, opts, interactive)
	if err != nil {
		return err
	}

	cfg := generateWorkspaceManifest(alias, name, outDir, port)
	manifestPath := filepath.Join(opts.path, manifestFileName)
	if err := writeOwnedFile(r, deps, manifestPath, opts.force, func() error {
		return writeWorkspaceManifest(manifestPath, cfg)
	}); err != nil {
		return err
	}

	guidePath := filepath.Join(opts.path, genericGuideFileName)
	if err := writeOwnedFile(r, deps, guidePath, opts.force, func() error {
		return deps.fetchGuide(guidePath)
	}); err != nil {
		return err
	}

	fmt.Fprintf(deps.out, "\nPrepared %q as a UI plugin workspace.\n\nNext steps:\n  Review %s and %s\n  Install @sailpoint/ui-plugin-sdk per the guide\n  sail ui-plugins create\n", opts.path, manifestFileName, genericGuideFileName)
	return nil
}

func resolveName(r *bufio.Reader, deps initDeps, opts initOptions, interactive bool) (string, error) {
	if opts.name != "" {
		return opts.name, nil
	}
	if !interactive {
		return "", fmt.Errorf("plugin name is required (set --name)")
	}
	v, err := promptLine(r, deps.out, "Plugin Name", "")
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(v) == "" {
		return "", fmt.Errorf("plugin name is required")
	}
	return strings.TrimSpace(v), nil
}

// resolveAndCheckAlias resolves the alias (an --alias value, or a slug of the
// name) and validates it against UMS. In interactive mode a definitive negative
// — already taken (409) or invalid (400) — reports why and re-prompts for a new
// alias (the correctable case) rather than aborting the whole command; headless
// mode fails fast. An inconclusive check (no client / unreachable) warns and is
// accepted, matching create's best-effort behavior.
func resolveAndCheckAlias(ctx context.Context, r *bufio.Reader, deps initDeps, opts initOptions, name string, interactive bool) (string, error) {
	def := slugifyAlias(name)

	if !interactive {
		alias := opts.alias
		if alias == "" {
			if def == "" {
				return "", fmt.Errorf("could not derive an alias from %q; set --alias", name)
			}
			alias = def
		}
		if err := checkAlias(ctx, deps.client, deps.errOut, alias); err != nil {
			return "", err
		}
		return alias, nil
	}

	// Interactive: on a definitive negative, report why and prompt again. The
	// attempt cap prevents an infinite loop when stdin is exhausted.
	alias := opts.alias
	for attempt := 0; attempt < maxAliasPrompts; attempt++ {
		if alias == "" {
			v, err := promptLine(r, deps.out, "Plugin Alias", def)
			if err != nil {
				return "", err
			}
			alias = strings.TrimSpace(v)
			if alias == "" {
				return "", fmt.Errorf("plugin alias is required")
			}
		}
		err := checkAlias(ctx, deps.client, deps.errOut, alias)
		if err == nil {
			return alias, nil
		}
		fmt.Fprintf(deps.out, "%v — please choose a different alias.\n", err)
		alias = ""
	}
	return "", fmt.Errorf("no valid alias provided after %d attempts", maxAliasPrompts)
}

func resolveOutDir(r *bufio.Reader, deps initDeps, opts initOptions, interactive bool) (string, error) {
	outDir := opts.outDir
	if outDir == "" {
		if !interactive {
			return "", fmt.Errorf("--out-dir is required when attaching with --path")
		}
		v, err := promptLine(r, deps.out, "Build Output Directory", "")
		if err != nil {
			return "", err
		}
		outDir = strings.TrimSpace(v)
	}
	if outDir == "" {
		return "", fmt.Errorf("build output directory is required")
	}
	// Reject NUL and other control characters — invalid in filesystem paths on
	// every platform and a sign of a corrupt/mis-pasted value. Existence and
	// "is this a real build directory" are validated later by update/upload
	// since the directory may not exist until the build runs.
	if hasControlChars(outDir) {
		return "", fmt.Errorf("build output directory contains invalid control characters")
	}
	return outDir, nil
}

// hasControlChars reports whether s contains any control character (including
// NUL), which are not valid in filesystem paths.
func hasControlChars(s string) bool {
	for _, r := range s {
		if unicode.IsControl(r) {
			return true
		}
	}
	return false
}

func resolvePort(r *bufio.Reader, deps initDeps, opts initOptions, interactive bool) (int, error) {
	if opts.portSet {
		if opts.port <= 0 {
			return 0, fmt.Errorf("port must be greater than 0")
		}
		return opts.port, nil
	}
	if !interactive {
		return defaultDevServerPort, nil
	}
	v, err := promptLine(r, deps.out, "Port", strconv.Itoa(defaultDevServerPort))
	if err != nil {
		return 0, err
	}
	port, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil {
		return 0, fmt.Errorf("port must be a number: %q", strings.TrimSpace(v))
	}
	if port <= 0 {
		return 0, fmt.Errorf("port must be greater than 0")
	}
	return port, nil
}

// writeOwnedFile writes a file init manages, prompting before overwriting an
// existing one unless --force is set. Files init does not manage are never
// touched.
func writeOwnedFile(r *bufio.Reader, deps initDeps, path string, force bool, write func() error) error {
	if _, err := os.Stat(path); err == nil && !force {
		overwrite, err := promptYesNoShared(r, deps.out, fmt.Sprintf("%s already exists. Overwrite?", filepath.Base(path)))
		if err != nil {
			return err
		}
		if !overwrite {
			fmt.Fprintf(deps.out, "Skipped %s\n", filepath.Base(path))
			return nil
		}
	}
	return write()
}

// checkAlias validates the alias against the UMS validate-alias endpoint. A
// definitive negative — already taken (409) or invalid (400) — fails the
// command. Any inconclusive result (no client, transport error, or another
// non-2xx status such as an auth/route problem) is reported as a warning and
// allowed to continue, so a misconfigured or unreachable backend doesn't block
// getting to a working local workspace.
func checkAlias(ctx context.Context, c client.Client, errOut io.Writer, alias string) error {
	if c == nil {
		fmt.Fprintf(errOut, "Warning: alias %q was not validated (no authenticated client). %s\n", alias, aliasNotValidatedHint)
		return nil
	}
	url := validateAliasEndpoint + "?alias=" + neturl.QueryEscape(alias)
	resp, err := c.Get(ctx, url, uiPluginRequestHeaders())
	if err != nil {
		fmt.Fprintf(errOut, "Warning: alias %q was not validated (%v). %s\n", alias, err, aliasNotValidatedHint)
		return nil
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	switch {
	case resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices:
		return nil
	case resp.StatusCode == http.StatusConflict:
		return fmt.Errorf("alias %q is already in use in this tenant", alias)
	case resp.StatusCode == http.StatusBadRequest:
		return fmt.Errorf("alias %q is invalid: %s", alias, umsErrorMessage(body))
	default:
		fmt.Fprintf(errOut, "Warning: alias %q could not be validated (status %d): %s. %s\n", alias, resp.StatusCode, umsErrorMessage(body), aliasNotValidatedHint)
		return nil
	}
}

// aliasNotValidatedHint is appended to best-effort validation warnings so the
// user knows the alias is re-checked later and how to resolve an unexpected miss.
const aliasNotValidatedHint = "The alias will be checked again by `sail ui-plugins create`; fix your CLI authentication if this is unexpected."

// defaultScaffold fetches the templates repo archive and extracts the Angular
// starter subtree into destDir (no template substitution — personalization
// happens afterward).
func defaultScaffold(destDir string) error {
	body, err := initialize.FetchArchive(uiPluginTemplatesRepoOwner, uiPluginTemplatesRepoName, uiPluginTemplatesBranch)
	if err != nil {
		return err
	}
	defer body.Close()
	return initialize.ExtractSubtree(body, destDir, angularStarterSubdir)
}

// defaultFetchGuide downloads the generic plugin guide from the templates repo
// and writes it to destPath.
func defaultFetchGuide(destPath string) error {
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", uiPluginTemplatesRepoOwner, uiPluginTemplatesRepoName, uiPluginTemplatesBranch, genericGuideFileName)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch plugin guide: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch plugin guide: HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read plugin guide: %w", err)
	}
	return os.WriteFile(destPath, data, 0644)
}

func promptLine(r *bufio.Reader, out io.Writer, label, def string) (string, error) {
	if def != "" {
		fmt.Fprintf(out, "%s [%s]: ", label, def)
	} else {
		fmt.Fprintf(out, "%s: ", label)
	}
	line, err := r.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	v := strings.TrimSpace(line)
	if v == "" {
		return def, nil
	}
	return v, nil
}

func promptYesNoShared(r *bufio.Reader, out io.Writer, question string) (bool, error) {
	fmt.Fprintf(out, "%s [y/N]: ", question)
	line, err := r.ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes", nil
}
