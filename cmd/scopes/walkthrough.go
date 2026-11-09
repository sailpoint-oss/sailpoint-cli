package scopes

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/sailpoint-oss/sailpoint-cli/cmd/ui_plugins"
	scopeanalysis "github.com/sailpoint-oss/sailpoint-cli/internal/scopes"
	"github.com/sailpoint-oss/sailpoint-cli/internal/tui"
)

// maxFolderPrompts bounds the folder re-prompt loop so exhausted stdin cannot
// loop forever.
const maxFolderPrompts = 3

// errCancelled is returned when the user leaves a list without a choice.
var errCancelled = errors.New("cancelled")

// prompter asks the walkthrough questions. Tests replace it with a fake.
type prompter interface {
	// Select returns the Id of the chosen item, or errCancelled.
	Select(title string, choices []tui.Choice) (string, error)
	Input(label string) (string, error)
	Confirm(question string) (bool, error)
}

// terminalPrompter shows lists with the CLI's bubbletea list and reads typed
// answers from stdin. Questions go to out, which is stderr, so stdout only
// holds the report.
type terminalPrompter struct {
	r   *bufio.Reader
	out io.Writer
}

func newTerminalPrompter(out io.Writer) *terminalPrompter {
	return &terminalPrompter{r: bufio.NewReader(os.Stdin), out: out}
}

func (p *terminalPrompter) Select(title string, choices []tui.Choice) (string, error) {
	choice, err := tui.PromptList(choices, title)
	if err != nil {
		return "", err
	}
	if choice.Id == "" {
		return "", errCancelled
	}
	_, _ = fmt.Fprintf(p.out, "%s %s\n", title+":", choice.Title)
	return choice.Id, nil
}

func (p *terminalPrompter) Input(label string) (string, error) {
	_, _ = fmt.Fprintf(p.out, "%s: ", label)
	line, err := p.r.ReadString('\n')
	if err != nil && (err != io.EOF || line == "") {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func (p *terminalPrompter) Confirm(question string) (bool, error) {
	answer, err := p.Input(question + " [y/N]")
	if err != nil {
		return false, err
	}
	answer = strings.ToLower(answer)
	return answer == "y" || answer == "yes", nil
}

// askAuthType asks for the token type unless --auth-type was passed.
func askAuthType(p prompter, opts *analyzeOptions) error {
	if opts.authTypeSet {
		return nil
	}
	id, err := p.Select("Which token will call the APIs?", []tui.Choice{
		{Id: string(scopeanalysis.AuthUser), Title: "User token", Description: "Personal access token, or a UI plugin running as the signed-in user"},
		{Id: string(scopeanalysis.AuthApplication), Title: "Application token", Description: "Client credentials grant, no user"},
	})
	if err != nil {
		return err
	}
	opts.authType = id
	return nil
}

// askFolder asks which folder to scan unless --path was passed.
func askFolder(p prompter, opts *analyzeOptions, errOut io.Writer) error {
	if opts.pathSet {
		return nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	id, err := p.Select("Which folder do you want to scan?", []tui.Choice{
		{Id: "cwd", Title: "This folder", Description: cwd},
		{Id: "other", Title: "A different folder", Description: "Type the path of the folder"},
	})
	if err != nil {
		return err
	}
	if id == "cwd" {
		opts.path = "."
		return nil
	}

	for range maxFolderPrompts {
		dir, err := p.Input("Folder to scan")
		if err != nil {
			return err
		}
		if dir == "" {
			continue
		}
		if info, err := os.Stat(dir); err != nil || !info.IsDir() {
			_, _ = fmt.Fprintf(errOut, "%s is not a folder.\n", dir)
			continue
		}
		opts.path = dir
		// Another folder gets absolute locations, the same as --path.
		opts.absolutePaths = true
		return nil
	}
	return fmt.Errorf("no folder to scan after %d attempts", maxFolderPrompts)
}

// offerManifestUpdate handles a sp-ui-plugin.json in the scanned folder when
// --update-manifest was not passed. In the walkthrough it shows the change and
// asks before writing. Otherwise it prints how to write the scopes.
func offerManifestUpdate(opts analyzeOptions, res *scopeanalysis.Result, errOut io.Writer) error {
	path := filepath.Join(opts.path, ui_plugins.ManifestFileName)
	if _, err := os.Stat(path); err != nil || len(res.RequiredScopes) == 0 {
		return nil
	}

	current, err := ui_plugins.ReadManifestAPIScopes(path)
	if err != nil {
		_, _ = fmt.Fprintf(errOut, "\nFound %s, but it cannot be updated: %v\n", path, err)
		return nil
	}
	if slices.Equal(current, res.RequiredScopes) {
		_, _ = fmt.Fprintf(errOut, "\napiScopes in %s already match the required scopes.\n", path)
		return nil
	}

	if !opts.interactive {
		_, _ = fmt.Fprintf(errOut, "\nFound %s. To write these scopes into apiScopes, run the command again with --update-manifest.\n", path)
		return nil
	}

	_, _ = fmt.Fprintf(errOut, "\nFound %s.\n", path)
	_, _ = fmt.Fprintf(errOut, "  Current apiScopes: %s\n", describeScopeList(current))
	_, _ = fmt.Fprintf(errOut, "  New apiScopes:     %s\n", describeScopeList(res.RequiredScopes))
	ok, err := opts.prompt.Confirm("Update apiScopes to the least privilege scopes?")
	if err != nil {
		return err
	}
	if !ok {
		_, _ = fmt.Fprintf(errOut, "%s was not changed.\n", path)
		return nil
	}
	if err := ui_plugins.UpdateManifestAPIScopes(path, res.RequiredScopes); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(errOut, "Updated apiScopes in %s.\n", path)
	return nil
}

func describeScopeList(scopes []string) string {
	if len(scopes) == 0 {
		return "(none)"
	}
	return strings.Join(scopes, ", ")
}
