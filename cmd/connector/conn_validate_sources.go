// Copyright (c) 2022, SailPoint Technologies, Inc. All rights reserved.
package connector

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/logrusorgru/aurora"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	connvalidate "github.com/sailpoint-oss/sailpoint-cli/cmd/connector/validate"
	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
)

type Source struct {
	// Name represents name of a source (github, smartsheet, freshservice, etc)
	Name string `yaml:"name"`
	// Repository is a link for a connector repository
	Repository string `yaml:"repository"`
	// RepositoryRef is the revision that uses for service starts. A full commit
	// SHA is required unless mutable revisions are explicitly allowed.
	RepositoryRef string `yaml:"repositoryRef"`
	// Config is an authentication data for service startup
	Config string `yaml:"config"`
	// ReadOnly is a flag that indicates the validation checks with data modification ('true') or without it ('false').
	ReadOnly bool `yaml:"readOnly"`
}

// ValidationResults represents validation results for every source
type ValidationResults struct {
	sourceName string
	results    map[string]*tablewriter.Table
}

const (
	connectorInstanceEndpoint = "http://localhost:3000"
	sourceFile                = "./source.yaml"
	// instanceStartTimeout bounds how long a connector instance gets to answer
	// on its endpoint before validation gives up and cleans the checkout up.
	instanceStartTimeout = 2 * time.Minute
	// allowedRepositoryHost is the only host validate-sources clones from.
	allowedRepositoryHost = "github.com"
)

// defaultAllowedRepositoryOwners lists the repository owners whose connectors
// validate-sources trusts enough to clone and execute by default.
var defaultAllowedRepositoryOwners = []string{"sailpoint", "sailpoint-oss"}

var (
	// commitSHAPattern matches a full, immutable git commit digest.
	commitSHAPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)
	// gitRefPattern limits mutable revisions to the characters git accepts in
	// branch and tag names, and keeps option-like values out of git arguments.
	gitRefPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/-]*$`)
	// repositoryPathPattern matches "<owner>/<repo>" with an optional .git suffix.
	repositoryPathPattern = regexp.MustCompile(`^([A-Za-z0-9](?:[A-Za-z0-9-]*[A-Za-z0-9])?)/([A-Za-z0-9._-]+?)(?:\.git)?$`)
	// scpLikePattern matches the "git@host:owner/repo" remote form.
	scpLikePattern = regexp.MustCompile(`^git@[A-Za-z0-9.-]+:[^:]+$`)
)

// sourceTrustPolicy decides which source entries validate-sources is willing to
// clone and execute.
type sourceTrustPolicy struct {
	allowedOwners   []string
	allowMutableRef bool
}

func (v *ValidationResults) Render() {
	fmt.Println(aurora.Blue(fmt.Sprintf("%s connectors validation results", v.sourceName)).String())
	for connectorID, result := range v.results {
		fmt.Println(aurora.Blue(fmt.Sprintf("Connector %s", connectorID)).String())
		result.Render()
		fmt.Println("---------------------------------------------------------")
	}
}

func newConnValidateSourcesCmd(apiClient client.Client) *cobra.Command {
	var (
		policy sourceTrustPolicy
		force  bool
	)

	cmd := &cobra.Command{
		Use:   "validate-sources",
		Short: "Validate connectors behavior",
		Long: `Validate connectors behavior from a list that stores in source.yaml.

Every listed repository is cloned and its dev script runs with your identity and
credentials, so each entry must name an allowed repository owner on ` + allowedRepositoryHost + ` and
pin a full commit SHA. Use --allow-mutable-ref to accept a branch or tag instead.`,
		Example: "sail conn validate-sources",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			endpoint := cmd.Flags().Lookup("conn-endpoint").Value.String()
			readLimitVal, err := getReadLimitVal(cmd)
			if err != nil {
				return fmt.Errorf("invalid value of readLimit: %v", err)
			}

			listOfSources, err := getSourceFromFile(sourceFile)
			if err != nil {
				return err
			}

			if len(listOfSources) == 0 {
				return fmt.Errorf("no sources are listed in %s", sourceFile)
			}

			// Reject every untrusted entry before anything is cloned or executed.
			if err := checkSourcesTrusted(listOfSources, policy); err != nil {
				return err
			}

			if !force {
				confirmed, err := confirmSources(cmd.InOrStdin(), cmd.ErrOrStderr(), listOfSources)
				if err != nil {
					return err
				}
				if !confirmed {
					return errors.New("validation cancelled")
				}
			}

			var results []ValidationResults

			for _, source := range listOfSources {
				res, err := runSourceValidation(ctx, apiClient, source, endpoint, readLimitVal)
				if err != nil {
					return err
				}

				results = append(results, *res)
			}

			for _, r := range results {
				r.Render()
			}

			return nil
		},
	}

	cmd.Flags().StringSliceVar(&policy.allowedOwners, "allowed-owner", defaultAllowedRepositoryOwners, "Repository owners that validate-sources is allowed to clone and execute")
	cmd.Flags().BoolVar(&policy.allowMutableRef, "allow-mutable-ref", false, "Accept a branch or tag revision instead of requiring a pinned commit SHA")
	cmd.Flags().BoolVarP(&force, "force", "F", false, "Bypass confirmation prompts")

	return cmd
}

func getSourceFromFile(filePath string) ([]Source, error) {
	yamlFile, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config []Source

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return nil, err
	}

	return config, err
}

// checkSourcesTrusted reports the first source entry that the policy rejects.
func checkSourcesTrusted(sources []Source, policy sourceTrustPolicy) error {
	for i, source := range sources {
		if err := checkSourceTrusted(source, policy); err != nil {
			name := source.Name
			if strings.TrimSpace(name) == "" {
				name = fmt.Sprintf("entry %d", i+1)
			}
			return fmt.Errorf("source %s in %s is not allowed: %w", name, sourceFile, err)
		}
	}

	return nil
}

// checkSourceTrusted verifies that a source names an allowed repository owner and
// a revision the policy accepts.
func checkSourceTrusted(source Source, policy sourceTrustPolicy) error {
	if strings.TrimSpace(source.Name) == "" {
		return errors.New("name is empty")
	}

	owner, err := parseRepositoryOwner(source.Repository)
	if err != nil {
		return err
	}

	if !ownerAllowed(owner, policy.allowedOwners) {
		return fmt.Errorf("repository owner %q is not in the allowed owners (%s); pass --allowed-owner to add it", owner, strings.Join(policy.allowedOwners, ", "))
	}

	return checkRepositoryRef(source.RepositoryRef, policy.allowMutableRef)
}

// parseRepositoryOwner returns the owner of a repository remote, rejecting any
// remote whose transport or host validate-sources does not trust. Only https,
// ssh, and the "git@host:owner/repo" form on the allowed host are accepted, which
// keeps local paths, option-like values, and git remote helpers such as "ext::"
// out of the git command line.
func parseRepositoryOwner(repository string) (string, error) {
	repo := strings.TrimSpace(repository)
	if repo == "" {
		return "", errors.New("repository is empty")
	}

	if strings.HasPrefix(repo, "-") {
		return "", fmt.Errorf("repository %q cannot start with '-'", repo)
	}

	var host, path string

	switch {
	case strings.HasPrefix(repo, "https://"), strings.HasPrefix(repo, "ssh://"):
		parsed, err := url.Parse(repo)
		if err != nil {
			return "", fmt.Errorf("repository %q is not a valid URL: %w", repo, err)
		}
		if parsed.User != nil && parsed.User.Username() != "git" {
			return "", fmt.Errorf("repository %q must not embed credentials", repo)
		}
		host = parsed.Hostname()
		path = strings.TrimPrefix(parsed.Path, "/")
	case scpLikePattern.MatchString(repo):
		hostAndPath := strings.SplitN(strings.TrimPrefix(repo, "git@"), ":", 2)
		host, path = hostAndPath[0], hostAndPath[1]
	default:
		return "", fmt.Errorf("repository %q must be an https://, ssh://, or git@%s:<owner>/<repo> remote", repo, allowedRepositoryHost)
	}

	if !strings.EqualFold(host, allowedRepositoryHost) {
		return "", fmt.Errorf("repository %q must be hosted on %s", repo, allowedRepositoryHost)
	}

	match := repositoryPathPattern.FindStringSubmatch(path)
	if match == nil {
		return "", fmt.Errorf("repository %q must name a single <owner>/<repo> path", repo)
	}

	return match[1], nil
}

func ownerAllowed(owner string, allowedOwners []string) bool {
	for _, allowed := range allowedOwners {
		if strings.EqualFold(strings.TrimSpace(allowed), owner) {
			return true
		}
	}

	return false
}

// checkRepositoryRef requires an immutable commit digest unless mutable
// revisions are allowed, in which case the branch or tag name still has to be one
// git accepts and one that cannot be read as a git option.
func checkRepositoryRef(repositoryRef string, allowMutableRef bool) error {
	ref := strings.TrimSpace(repositoryRef)
	if ref == "" {
		return errors.New("repositoryRef is empty; set it to a full 40 character commit SHA")
	}

	if commitSHAPattern.MatchString(ref) {
		return nil
	}

	if !allowMutableRef {
		return fmt.Errorf("repositoryRef %q is not a full 40 character commit SHA; pin the revision or pass --allow-mutable-ref", ref)
	}

	if !gitRefPattern.MatchString(ref) ||
		strings.Contains(ref, "..") ||
		strings.Contains(ref, "@{") ||
		strings.HasSuffix(ref, ".lock") ||
		strings.HasSuffix(ref, "/") {
		return fmt.Errorf("repositoryRef %q is not a valid branch or tag name", ref)
	}

	return nil
}

// confirmSources shows what is about to be cloned and executed and waits for an
// explicit yes. An empty answer, or no answer at all, cancels validation.
func confirmSources(in io.Reader, w io.Writer, sources []Source) (bool, error) {
	fmt.Fprintf(w, "validate-sources clones these repositories and runs their dev script with your identity and credentials:\n")
	for _, source := range sources {
		fmt.Fprintf(w, "  - %s: %s at %s\n", source.Name, strings.TrimSpace(source.Repository), strings.TrimSpace(source.RepositoryRef))
	}

	fmt.Fprint(w, "Continue? [y/N]: ")
	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}

	answer := strings.ToLower(strings.TrimSpace(line))

	return answer == "y" || answer == "yes", nil
}

// runSourceValidation starts one connector instance, validates it, and always
// stops the instance and removes its checkout before returning.
func runSourceValidation(ctx context.Context, apiClient client.Client, source Source, endpoint string, readLimit int64) (res *ValidationResults, err error) {
	instance, tempFolder, err := runInstanceForValidation(source)
	if err != nil {
		return nil, err
	}

	defer func() {
		if stopErr := util.StopCommand(instance); stopErr != nil && err == nil {
			err = fmt.Errorf("%s instance wasn't stopped: %w", source.Name, stopErr)
		}

		if removeErr := os.RemoveAll(tempFolder); removeErr != nil && err == nil {
			err = removeErr
		}
	}()

	return validateConnectors(ctx, apiClient, source, endpoint, readLimit)
}

func validateConnectors(ctx context.Context, apiClient client.Client, source Source, endpoint string, readLimit int64) (*ValidationResults, error) {
	resp, err := apiClient.Get(ctx, endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("non-200 response for getting all %s connectors: %s\nbody: %s", source.Name, resp.Status, body)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var conns []connector
	err = json.Unmarshal(raw, &conns)
	if err != nil {
		return nil, err
	}

	valRes := &ValidationResults{
		sourceName: source.Name,
		results:    make(map[string]*tablewriter.Table),
	}

	connector := conns[len(conns)-1]

	cc, err := connClientWithCustomParams(apiClient, json.RawMessage(source.Config), connector.ID, "0", connectorInstanceEndpoint)
	if err != nil {
		log.Println(err)
	}

	validator := connvalidate.NewValidator(connvalidate.Config{
		Check:     "",
		ReadOnly:  source.ReadOnly,
		ReadLimit: readLimit,
	}, cc)

	results, err := validator.Run(ctx)
	if err != nil {
		log.Println(err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]any{"ID", "Result", "Errors", "Warnings", "Skipped"}...)
	for _, res := range results {
		var result = aurora.Green("PASS")
		if len(res.Errors) > 0 {
			result = aurora.Red("FAIL")
		}

		if len(res.Skipped) > 0 {
			result = aurora.Yellow("SKIPPED")
		}

		table.Append([]string{
			aurora.Blue(res.ID).String(),
			result.String(),
			aurora.Red(strings.Join(res.Errors, "\n\n")).String(),
			aurora.Yellow(strings.Join(res.Warnings, "\n\n")).String(),
			aurora.Yellow(strings.Join(res.Skipped, "\n\n")).String(),
		})
	}

	valRes.results[connector.ID] = table

	return valRes, err
}

func createTempFolder() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	path, err := os.MkdirTemp(homeDir, "*")
	if err != nil {
		return "", err
	}

	return path, err
}

func runInstanceForValidation(source Source) (*exec.Cmd, string, error) {
	path, err := createTempFolder()
	if err != nil {
		return nil, "", err
	}

	repository := strings.TrimSpace(source.Repository)
	repositoryRef := strings.TrimSpace(source.RepositoryRef)

	// Every git and npm call below runs without a shell, so the source values
	// stay single arguments instead of becoming shell syntax.
	cloneRepo := exec.Command("git", "clone", "--", repository, path)
	if err := cloneRepo.Run(); err != nil {
		return nil, "", err
	}

	log.Printf("Repo for %s is cloned\n", source.Name)

	checkoutRepoRef := exec.Command("git", "-C", path, "checkout", repositoryRef)
	if err := checkoutRepoRef.Run(); err != nil {
		return nil, "", err
	}

	log.Printf("git checkout to %s\n", repositoryRef)

	// --ignore-scripts keeps dependency lifecycle hooks from running during
	// installation. Only the dev script the command needs is executed.
	install := exec.Command("npm", "install", "--ignore-scripts", "--prefix", path)
	if err := install.Run(); err != nil {
		return nil, "", err
	}

	log.Println("Npm install is finished")

	instance, err := util.ExecCommand("npm", "run", "dev", "--prefix", path)
	if err != nil {
		return nil, "", err
	}

	if err := waitForInstance(instance, source.Name); err != nil {
		if stopErr := util.StopCommand(instance); stopErr != nil {
			log.Println(stopErr)
		}
		return nil, "", err
	}

	return instance, path, nil
}

// waitForInstance polls the connector endpoint until the instance answers, and
// gives up once instanceStartTimeout passes so a connector that never starts
// does not hang the command.
func waitForInstance(instance *exec.Cmd, sourceName string) error {
	deadline := time.Now().Add(instanceStartTimeout)

	for {
		resp, err := http.Get(connectorInstanceEndpoint)
		if err == nil {
			_ = resp.Body.Close()
			log.Printf("Service %s is successfully started for validation\n", sourceName)
			return nil
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("service %s did not answer on %s within %s", sourceName, connectorInstanceEndpoint, instanceStartTimeout)
		}

		time.Sleep(time.Second * 5)
	}
}
