// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package reassign

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	sailpoint "github.com/sailpoint-oss/golang-sdk/v2"
	"github.com/sailpoint-oss/golang-sdk/v2/api_v2024"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

//go:embed reassign.md
var reassignHelp string

var supportedObjectTypes = []string{
	"source",
	"role",
	"access-profile",
	"entitlement",
	"identity-profile",
	"governance-group",
	"workflow",
}

type Identity struct {
	ID   string
	Name string
}

type ReassignSummary struct {
	From             Identity
	To               Identity
	ObjectCounts     map[string]int
	ObjectTypes      []string
	DryRun           bool
	Verbose          bool
	Sources          []api_v2024.Source
	Roles            []api_v2024.Role
	AccessProfiles   []api_v2024.AccessProfile
	Entitlements     []api_v2024.Entitlement
	IdentityProfiles []api_v2024.IdentityProfile
	GovernanceGroups []api_v2024.WorkgroupDto
	Workflows        []api_v2024.Workflow
}

func (r ReassignSummary) IsEmpty() bool {
	return len(r.Sources) == 0 &&
		len(r.Roles) == 0 &&
		len(r.AccessProfiles) == 0 &&
		len(r.Entitlements) == 0 &&
		len(r.IdentityProfiles) == 0 &&
		len(r.GovernanceGroups) == 0 &&
		len(r.Workflows) == 0
}

type errMsg error
type summaryMsg *ReassignSummary
type reassignDoneMsg struct{}
type statusMsg string

func NewReassignCommand() *cobra.Command {
	var from string
	var to string
	var objectTypes string
	var objectId string
	var dryRun bool
	var force bool

	help := util.ParseHelp(reassignHelp)
	cmd := &cobra.Command{
		Use:     "reassign",
		Short:   "Reassign object ownership in Identity Security Cloud",
		Long:    help.Long,
		Example: help.Example,
		Args:    cobra.OnlyValidArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			if from == "" || to == "" {
				return errors.New("both --from and --to flags are required when using --object-id")
			}

			if dryRun && force {
				log.Error("cannot use --dry-run and --force together")
				os.Exit(1)
			}

			if from == to {
				log.Error("from and to Identities cannot be the same")
				os.Exit(1)
			}

			if objectId != "" {
				if from == "" || to == "" {
					return errors.New("both --from and --to flags are required when using --object-id")
				}

				summary, err := determineObjectTypeAndCreateReassignment(objectId, from, to, dryRun)

				if err != nil {
					log.Error("error determining object type:", "error", err)
					os.Exit(1)
				}

				if !force {
					printSummary(summary)

					if !summary.DryRun {
						promptSaveReport(&summary)

						fmt.Printf("Would you like to proceed with reassigning this object from '%s' to '%s': ", summary.From.Name, summary.To.Name)
						var reassignResponse string
						_, err = fmt.Scanln(&reassignResponse)
						if err != nil {
							fmt.Println("Failed to read input:", err)
							return err
						}

						response := strings.ToLower(strings.TrimSpace(reassignResponse))

						if response == "y" {
							m := initialModel(from, to, objectTypes, dryRun, force)
							m.reassignResult = &summary
							m.reassigning = true
							prog := tea.NewProgram(m)
							_, err = prog.Run()
							if err != nil {
								return err
							}

						} else {
							fmt.Println("Aborted reassignment.")
						}
					} else {
						promptSaveReport(&summary)
					}
				} else {
					m := initialModel(from, to, objectTypes, dryRun, force)
					m.reassignResult = &summary
					m.reassigning = true
					prog := tea.NewProgram(m)
					_, err = prog.Run()
					if err != nil {
						return err
					}
				}

				return nil
				//Determine the object type from the objectId
			}

			p := tea.NewProgram(initialModel(from, to, objectTypes, dryRun, force))
			finalModel, err := p.Run()
			if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}

			if m, ok := finalModel.(model); ok && m.err != nil {
				log.Error("An error occurred when gathering objects to reassign:", "error", m.err)
				os.Exit(1)
			}

			if m, ok := finalModel.(model); ok && m.reassignResult != nil {
				p.Quit()

				if m.reassignResult.IsEmpty() {
					fmt.Println("No objects to reassign.")
					return nil
				}

				if !m.force {
					printSummary(*m.reassignResult)

					// If this was not a dry run proceed with the reassignment flow
					if !m.reassignResult.DryRun {

						promptSaveReport(m.reassignResult)

						fmt.Printf("Would you like to proceed with reassigning these objects from '%s' to '%s': ", m.reassignResult.From.Name, m.reassignResult.To.Name)
						var reassignResponse string
						_, err = fmt.Scanln(&reassignResponse)
						if err != nil {
							fmt.Println("Failed to read input:", err)
							return err
						}

						response := strings.ToLower(strings.TrimSpace(reassignResponse))

						if response == "y" {
							m := initialModel(from, to, objectTypes, dryRun, force)
							m.reassignResult = finalModel.(model).reassignResult
							m.reassigning = true
							prog := tea.NewProgram(m)
							_, err = prog.Run()
							if err != nil {
								return err
							}

						} else {
							fmt.Println("Aborted reassignment.")
						}

					} else {
						promptSaveReport(m.reassignResult)
					}
				} else {
					m := initialModel(from, to, objectTypes, dryRun, force)
					m.reassignResult = finalModel.(model).reassignResult
					m.reassigning = true
					prog := tea.NewProgram(m)
					_, err = prog.Run()
					if err != nil {
						return err
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&from, "from", "f", "", "The identity to reassign from")
	cmd.Flags().StringVarP(&to, "to", "t", "", "The identity to reassign to")
	cmd.Flags().BoolVarP(&force, "force", "F", false, "Bypass confirmation prompts")
	cmd.Flags().StringVarP(&objectTypes, "object-types", "o", "", "Comma-separated list of object types to reassign, defaults to all")
	cmd.Flags().StringVarP(&objectId, "object-id", "i", "", "The object id to reassign")
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Show the objects that would be reassigned without actually reassigning them")

	return cmd

}

func determineObjectTypeAndCreateReassignment(objectId string, from string, to string, dryRun bool) (ReassignSummary, error) {
	var summary ReassignSummary
	var reassignIdentities []api_v2024.Identity
	var objectsToReassign []string

	apiClient, err := config.InitAPIClient(true)

	if err != nil {
		return summary, err
	}

	if from != "" && to != "" {
		filters := fmt.Sprintf("id in (\"%s\",\"%s\")", from, to)
		identities, _, err := apiClient.V2024.IdentitiesAPI.ListIdentities(context.TODO()).Filters(filters).Execute()
		if err != nil {
			return summary, err
		}
		if len(identities) != 2 {
			return summary, errors.New("unable to find identities with the provided IDs")
		}
		reassignIdentities = identities
	}

	var fromIdentity = Identity{
		ID:   from,
		Name: getNameByID(reassignIdentities, from),
	}
	var toIdentity = Identity{
		ID:   to,
		Name: getNameByID(reassignIdentities, to),
	}

	summary = NewReassignSummary(fromIdentity, toIdentity, objectsToReassign, dryRun)

	// Check if the objectId is a source
	source, resp, err := apiClient.V2024.SourcesAPI.GetSource(context.TODO(), objectId).Execute()
	if err != nil {
		if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusBadRequest {
			log.Debug("Error getting access profile:", "error", err)
		}
	} else {
		if resp.StatusCode == http.StatusOK {
			if source.Owner.Id != nil && *source.Owner.Id != from {
				return summary, errors.New("the source is not owned by the specified identity")
			}

			summary.Sources = append(summary.Sources, *source)
			summary.ObjectTypes = []string{"source"}
			summary.ObjectCounts["source"] = 1
			return summary, err
		}
	}

	// Check if the objectId is a role
	role, resp, err := apiClient.V2024.RolesAPI.GetRole(context.TODO(), objectId).Execute()
	if err != nil {
		if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusBadRequest {
			log.Debug("Error getting access profile:", "error", err)
		}
	} else {
		if resp.StatusCode == http.StatusOK {
			if role.Owner.Id != nil && *role.Owner.Id != from {
				return summary, errors.New("the role is not owned by the specified identity")
			}
			summary.Roles = append(summary.Roles, *role)
			summary.ObjectTypes = []string{"role"}
			summary.ObjectCounts["role"] = 1
			return summary, err
		}
	}

	// Check if the objectId is an access profile
	accessProfile, resp, err := apiClient.V2024.AccessProfilesAPI.GetAccessProfile(context.TODO(), objectId).Execute()
	if err != nil {
		if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusBadRequest {
			log.Debug("Error getting access profile:", "error", err)
		}
	} else {
		if resp.StatusCode == http.StatusOK {
			if accessProfile.Owner.Id != nil && *accessProfile.Owner.Id != from {
				return summary, errors.New("the access profile is not owned by the specified identity")
			}

			summary.AccessProfiles = append(summary.AccessProfiles, *accessProfile)
			summary.ObjectTypes = []string{"access-profile"}
			summary.ObjectCounts["access-profile"] = 1
			return summary, err
		}
	}

	entitlement, resp, err := apiClient.V2024.EntitlementsAPI.GetEntitlement(context.TODO(), objectId).Execute()
	if err != nil {
		if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusBadRequest {
			log.Debug("Error getting entitlement:", "error", err)
		}
	} else {
		if resp.StatusCode == http.StatusOK {
			if entitlement.Owner.Get().Id != nil && *entitlement.Owner.Get().Id != from {
				return summary, errors.New("the entitlement is not owned by the specified identity")
			}
			summary.Entitlements = append(summary.Entitlements, *entitlement)
			summary.ObjectTypes = []string{"entitlement"}
			summary.ObjectCounts["entitlement"] = 1
			return summary, err
		}
	}

	identityProfile, resp, err := apiClient.V2024.IdentityProfilesAPI.GetIdentityProfile(context.TODO(), objectId).Execute()
	if err != nil {
		if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusBadRequest {
			log.Debug("Error getting identity profile:", "error", err)
		}
	} else {
		if resp.StatusCode == http.StatusOK {
			if identityProfile.Owner.Get().Id != nil && *identityProfile.Owner.Get().Id != from {
				return summary, errors.New("the identity profile is not owned by the specified identity")
			}

			summary.IdentityProfiles = append(summary.IdentityProfiles, *identityProfile)
			summary.ObjectTypes = []string{"identity-profile"}
			summary.ObjectCounts["identity-profile"] = 1
			return summary, err
		}
	}

	governanceGroup, resp, err := apiClient.V2024.GovernanceGroupsAPI.GetWorkgroup(context.TODO(), objectId).Execute()
	if err != nil {
		if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusBadRequest {
			log.Debug("Error getting governance group:", "error", err)
		}
	} else {
		if resp.StatusCode == http.StatusOK {
			if governanceGroup.Owner.Id != nil && *governanceGroup.Owner.Id != from {
				return summary, errors.New("the governance group is not owned by the specified identity")
			}

			summary.GovernanceGroups = append(summary.GovernanceGroups, *governanceGroup)
			summary.ObjectTypes = []string{"governance-group"}
			summary.ObjectCounts["governance-group"] = 1
			return summary, err
		}
	}

	workflow, resp, err := apiClient.V2024.WorkflowsAPI.GetWorkflow(context.TODO(), objectId).Execute()
	if err != nil {
		if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusBadRequest {
			log.Debug("Error getting workflow:", "error", err)
		}
	} else {
		if resp.StatusCode == http.StatusOK {
			if workflow.Owner.Id != nil && *workflow.Owner.Id != from {
				return summary, errors.New("the workflow is not owned by the specified identity")
			}

			summary.Workflows = append(summary.Workflows, *workflow)
			summary.ObjectTypes = []string{"workflow"}
			summary.ObjectCounts["workflow"] = 1
			return summary, err
		}
	}

	return summary, errors.New("object not found")
}

func promptSaveReport(summary *ReassignSummary) error {
	fmt.Print("Would you like to save the full report to a file (y/n): ")
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}
	response = strings.ToLower(strings.TrimSpace(response))
	if response != "y" {
		return nil
	}

	fmt.Print("Enter the file name (without extension)(default: reassign_report): ")
	var fileName string
	_, err = fmt.Scanln(&fileName)
	if err != nil && err.Error() != "unexpected newline" {
		return fmt.Errorf("failed to read input: %w", err)
	}
	if strings.TrimSpace(fileName) == "" {
		fileName = "reassign_report"
	}
	return writeReport(*summary, fmt.Sprintf("%s.json", fileName))
}

func writeReport(summary ReassignSummary, path string) error {

	file, err := os.Create(path)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.WriteString(util.PrettyPrint(summary))
	if err != nil {
		return err
	}

	err = file.Sync()
	if err != nil {
		return err
	}

	fmt.Print("Report saved to ", path, "\n")

	return nil
}

func isValidType(value string) bool {
	for _, t := range supportedObjectTypes {
		if value == t {
			return true
		}
	}
	return false
}

func validateObjectTypes(input string) error {
	types := strings.Split(input, ",")
	for _, t := range types {
		t = strings.TrimSpace(t)
		if !isValidType(t) {
			return fmt.Errorf("unsupported object type: '%s'", t)
		}
	}
	return nil
}

func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

func printSummary(summary ReassignSummary) {
	fmt.Println("Reassignment Preview")
	fmt.Println("====================")
	fmt.Printf("From Owner:       %s (%s)\n", summary.From.ID, summary.From.Name)
	fmt.Printf("To Owner:         %s (%s)\n", summary.To.ID, summary.To.Name)

	fmt.Printf("Object Types:     %s\n", strings.Join(summary.ObjectTypes, ", "))
	fmt.Printf("Dry Run:          %t\n\n", summary.DryRun)

	fmt.Println("Objects to Reassign:")
	fmt.Println("---------------------")
	fmt.Printf("%-20s %s\n", "Object Type", "Count")
	fmt.Printf("%-20s %s\n", "-----------", "-----")

	total := 0
	for objectType, count := range summary.ObjectCounts {
		fmt.Printf("%-20s %d\n", objectType, count)
		total += count
	}
	fmt.Printf("\nTotal:             %d objects\n\n", total)

	if summary.DryRun {
		fmt.Println("No changes have been made. Run the command without --dry-run to proceed.")
	}
}

func NewReassignSummary(fromIdentity Identity, toIdentity Identity, supportedObjectTypes []string, dryRun bool) ReassignSummary {
	return ReassignSummary{
		From:         fromIdentity,
		To:           toIdentity,
		DryRun:       dryRun,
		ObjectTypes:  supportedObjectTypes,
		ObjectCounts: make(map[string]int),
	}
}

func getNameByID(identities []api_v2024.Identity, targetID string) string {
	for _, identity := range identities {
		if *identity.Id == targetID {
			return identity.Name
		}
	}
	return "" // or return "not found" or an error
}

type model struct {
	spinner        spinner.Model
	quitting       bool
	from           string
	to             string
	objectTypes    string
	dryRun         bool
	force          bool
	err            error
	done           bool
	reassigning    bool
	statusText     string
	reassignResult *ReassignSummary
}

func initialModel(from string, to string, objectTypes string, dryRun bool, force bool) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return model{spinner: s, from: from, to: to, objectTypes: objectTypes, dryRun: dryRun, force: force}
}

func (m model) Init() tea.Cmd {
	if m.reassigning && m.reassignResult != nil {
		apiClient, err := config.InitAPIClient(true)
		if err != nil {
			return func() tea.Msg { return errMsg(err) }
		}
		return tea.Batch(m.spinner.Tick, nextReassignmentStepCmd(apiClient.V2024, *m.reassignResult, 0))
	}
	return tea.Batch(m.spinner.Tick, fetchReassignSummaryCmd(m.from, m.to, m.objectTypes, m.dryRun, m.force))
}

func fetchReassignSummaryCmd(from string, to string, objectTypes string, dryRun bool, force bool) tea.Cmd {
	return func() tea.Msg {
		// your logic here (init API, gather data, etc)
		// return errMsg(err) on error or summaryMsg(result)
		var objectsToReassign []string
		var reassignIdentities []api_v2024.Identity
		var sources []api_v2024.Source
		var roles []api_v2024.Role
		var accessProfiles []api_v2024.AccessProfile
		var identityProfiles []api_v2024.IdentityProfile
		var filteredIdentityProfiles []api_v2024.IdentityProfile
		var entitlements []api_v2024.Entitlement
		var governanceGroups []api_v2024.WorkgroupDto
		var filteredGovernanceGroups []api_v2024.WorkgroupDto
		var filteredWorkflows []api_v2024.Workflow
		var resp *http.Response

		apiClient, err := config.InitAPIClient(true)

		if err != nil {
			return errMsg(err)
		}
		if from != "" && to != "" {
			filters := fmt.Sprintf("id in (\"%s\",\"%s\")", from, to)
			identities, _, err := apiClient.V2024.IdentitiesAPI.ListIdentities(context.TODO()).Filters(filters).Execute()
			if err != nil {
				return errMsg(err)
			}
			if len(identities) != 2 {
				return errors.New("unable to find identities with the provided IDs")
			}
			reassignIdentities = identities
		}

		if objectTypes != "" {
			err := validateObjectTypes(objectTypes)
			if err != nil {
				return errMsg(err)
			}

			objectsToReassign = strings.Split(objectTypes, ",")
		} else {
			objectsToReassign = supportedObjectTypes
		}

		var fromIdentity = Identity{
			ID:   from,
			Name: getNameByID(reassignIdentities, from),
		}
		var toIdentity = Identity{
			ID:   to,
			Name: getNameByID(reassignIdentities, to),
		}

		var reassignSummary = NewReassignSummary(fromIdentity, toIdentity, objectsToReassign, dryRun)

		if contains(objectsToReassign, "source") {
			log.Debug("Gathering sources to reassign")

			filters := fmt.Sprintf("owner.id eq \"%s\"", from)
			_, resp, err = apiClient.V2024.SourcesAPI.ListSources(context.TODO()).Filters(filters).Count(true).Limit(1).Execute()

			if err != nil {
				return errMsg(fmt.Errorf("failed to get count of sources owned by id %s: %w", fromIdentity.ID, err))
			}

			count := resp.Header.Get("X-Total-Count")

			totalSources, err := strconv.Atoi(count)
			if err != nil {
				return errMsg(err)
			}

			if totalSources > 250 {
				// Paginate through the sources
				log.Debug("Paginating Sources, total to reassign: ", totalSources)
				sources, _, err = sailpoint.Paginate[api_v2024.Source](apiClient.V2024.SourcesAPI.ListSources(context.TODO()).Filters(filters), 0, 250, int32(totalSources))

				if err != nil {
					return errMsg(fmt.Errorf("failed to paginate sources owned by id %s: %w", fromIdentity.ID, err))
				}

			} else {
				sources, _, err = apiClient.V2024.SourcesAPI.ListSources(context.TODO()).Filters(filters).Execute()

				if err != nil {
					return errMsg(fmt.Errorf("failed to retrieve sources owned by id %s: %w", fromIdentity.ID, err))
				}
			}

			reassignSummary.Sources = sources
			reassignSummary.ObjectCounts["source"] = len(sources)

		}

		if contains(objectsToReassign, "role") {
			log.Debug("Gathering roles to reassign")
			filters := fmt.Sprintf("owner.id eq \"%s\"", from)
			_, resp, err = apiClient.V2024.RolesAPI.ListRoles(context.TODO()).Filters(filters).Count(true).Limit(1).Execute()
			if err != nil {
				return errMsg(fmt.Errorf("failed to get count of roles owned by id %s: %w", fromIdentity.ID, err))
			}

			count := resp.Header.Get("X-Total-Count")

			totalRoles, err := strconv.Atoi(count)
			if err != nil {
				return errMsg(err)
			}

			if totalRoles > 250 {
				// Paginate through the roles
				log.Debug("Paginating roles, total to reassign: ", totalRoles)
				roles, _, err = sailpoint.Paginate[api_v2024.Role](apiClient.V2024.RolesAPI.ListRoles(context.TODO()).Filters(filters), 0, 250, int32(totalRoles))

				if err != nil {
					return errMsg(fmt.Errorf("failed to paginate roles owned by id %s: %w", fromIdentity.ID, err))
				}

			} else {
				roles, _, err = apiClient.V2024.RolesAPI.ListRoles(context.TODO()).Filters(filters).Execute()

				if err != nil {
					return errMsg(fmt.Errorf("failed to retrieve roles owned by id %s: %w", fromIdentity.ID, err))
				}
			}

			reassignSummary.Roles = roles
			reassignSummary.ObjectCounts["role"] = len(roles)
		}

		if contains(objectsToReassign, "access-profile") {
			log.Debug("Gathering access profiles to reassign")
			filters := fmt.Sprintf("owner.id eq \"%s\"", from)
			_, resp, err = apiClient.V2024.AccessProfilesAPI.ListAccessProfiles(context.TODO()).Filters(filters).Count(true).Limit(1).Execute()

			if err != nil {
				return errMsg(fmt.Errorf("failed to get count of access profiles owned by id %s: %w", fromIdentity.ID, err))
			}

			count := resp.Header.Get("X-Total-Count")

			totalAccessProfiles, err := strconv.Atoi(count)
			if err != nil {
				return errMsg(err)
			}

			if totalAccessProfiles > 250 {
				// Paginate through the roles
				log.Debug("Paginating access profiles, total to reassign: ", totalAccessProfiles)
				accessProfiles, _, err = sailpoint.Paginate[api_v2024.AccessProfile](apiClient.V2024.AccessProfilesAPI.ListAccessProfiles(context.TODO()).Filters(filters), 0, 250, int32(totalAccessProfiles))

				if err != nil {
					return errMsg(fmt.Errorf("failed to paginate access profiles owned by id %s: %w", fromIdentity.ID, err))
				}

			} else {
				accessProfiles, _, err = apiClient.V2024.AccessProfilesAPI.ListAccessProfiles(context.TODO()).Filters(filters).Execute()

				if err != nil {
					return errMsg(fmt.Errorf("failed to retrieve access profiles owned by id %s: %w", fromIdentity.ID, err))
				}
			}

			reassignSummary.AccessProfiles = accessProfiles
			reassignSummary.ObjectCounts["access-profile"] = len(accessProfiles)
		}

		if contains(objectsToReassign, "entitlement") {
			log.Debug("Gathering entitlements to reassign")
			filters := fmt.Sprintf("owner.id eq \"%s\"", from)
			_, resp, err = apiClient.V2024.EntitlementsAPI.ListEntitlements(context.TODO()).Filters(filters).Count(true).Limit(1).Execute()
			if err != nil {
				return errMsg(fmt.Errorf("failed to get count of entitlements owned by id %s: %w", fromIdentity.ID, err))
			}

			count := resp.Header.Get("X-Total-Count")

			totalEntitlements, err := strconv.Atoi(count)
			if err != nil {
				return errMsg(err)
			}

			if totalEntitlements > 250 {
				// Paginate through the roles
				log.Debug("Paginating entitlements, total to reassign: ", totalEntitlements)
				entitlements, _, err = sailpoint.Paginate[api_v2024.Entitlement](apiClient.V2024.EntitlementsAPI.ListEntitlements(context.TODO()).Filters(filters), 0, 250, int32(totalEntitlements))

				if err != nil {
					return errMsg(fmt.Errorf("failed to paginate entitlements owned by id %s: %w", fromIdentity.ID, err))
				}

			} else {
				entitlements, _, err = apiClient.V2024.EntitlementsAPI.ListEntitlements(context.TODO()).Filters(filters).Execute()

				if err != nil {
					return errMsg(fmt.Errorf("failed to retrieve entitlements owned by id %s: %w", fromIdentity.ID, err))
				}
			}

			reassignSummary.Entitlements = entitlements
			reassignSummary.ObjectCounts["entitlement"] = len(entitlements)
		}

		if contains(objectsToReassign, "identity-profile") {
			log.Debug("Gathering identity profiles to reassign")
			_, resp, err := apiClient.V2024.IdentityProfilesAPI.ListIdentityProfiles(context.TODO()).Count(true).Limit(1).Execute()
			if err != nil {
				return errMsg(fmt.Errorf("failed to get count of identity profiles owned by id %s: %w", fromIdentity.ID, err))
			}

			count := resp.Header.Get("X-Total-Count")

			totalIdentityProfiles, err := strconv.Atoi(count)
			if err != nil {
				return errMsg(err)
			}

			if totalIdentityProfiles > 250 {
				// Paginate through the roles
				log.Debug("Paginating identity profiles, total to reassign: ", totalIdentityProfiles)
				identityProfiles, _, err = sailpoint.Paginate[api_v2024.IdentityProfile](apiClient.V2024.IdentityProfilesAPI.ListIdentityProfiles(context.TODO()), 0, 250, int32(totalIdentityProfiles))

				if err != nil {
					return errMsg(fmt.Errorf("failed to paginate identity profiles owned by id %s: %w", fromIdentity.ID, err))
				}

			} else {
				identityProfiles, _, err = apiClient.V2024.IdentityProfilesAPI.ListIdentityProfiles(context.TODO()).Execute()

				if err != nil {
					return errMsg(fmt.Errorf("failed to retrieve identity profiles owned by id %s: %w", fromIdentity.ID, err))
				}
			}

			// Filter identity profiles by owner
			for _, profile := range identityProfiles {
				if profile.Owner.Get() != nil {
					if *profile.Owner.Get().Id == from {
						filteredIdentityProfiles = append(filteredIdentityProfiles, profile)
					}
				}
			}

			reassignSummary.IdentityProfiles = filteredIdentityProfiles
			reassignSummary.ObjectCounts["identity-profile"] = len(filteredIdentityProfiles)
		}

		if contains(objectsToReassign, "governance-group") {
			log.Debug("Gathering governance groups to reassign")
			_, resp, err := apiClient.V2024.GovernanceGroupsAPI.ListWorkgroups(context.TODO()).Count(true).Limit(1).Execute()
			if err != nil {
				return errMsg(fmt.Errorf("failed to get count of governance groups owned by id %s: %w", fromIdentity.ID, err))
			}

			count := resp.Header.Get("X-Total-Count")

			totalGovernanceGroups, err := strconv.Atoi(count)
			if err != nil {
				return errMsg(err)
			}

			if totalGovernanceGroups > 250 {
				// Paginate through the roles
				log.Debug("Paginating governance groups, total to reassign: ", totalGovernanceGroups)
				governanceGroups, _, err = sailpoint.Paginate[api_v2024.WorkgroupDto](apiClient.V2024.GovernanceGroupsAPI.ListWorkgroups(context.TODO()), 0, 250, int32(totalGovernanceGroups))

				if err != nil {
					return errMsg(fmt.Errorf("failed to paginate governance groups owned by id %s: %w", fromIdentity.ID, err))
				}

			} else {
				governanceGroups, _, err = apiClient.V2024.GovernanceGroupsAPI.ListWorkgroups(context.TODO()).Execute()

				if err != nil {
					return errMsg(fmt.Errorf("failed to retrieve governance groups owned by id %s: %w", fromIdentity.ID, err))
				}
			}

			// Filter governance groups by owner
			for _, group := range governanceGroups {
				if group.Owner.Id != nil && *group.Owner.Id == from {
					filteredGovernanceGroups = append(filteredGovernanceGroups, group)
				}
			}

			reassignSummary.GovernanceGroups = filteredGovernanceGroups
			reassignSummary.ObjectCounts["governance-group"] = len(filteredGovernanceGroups)
		}

		// No need to paginate workflows due to the limit of 100 per tenant
		if contains(objectsToReassign, "workflow") {
			log.Debug("Gathering workflows to reassign")
			workflows, _, err := apiClient.V2024.WorkflowsAPI.ListWorkflows(context.TODO()).Execute()
			if err != nil {
				return errMsg(fmt.Errorf("failed to retrieve workflows owned by id %s: %w", fromIdentity.ID, err))
			}
			// Filter workflows by owner
			for _, workflow := range workflows {
				if workflow.Owner.Id != nil && *workflow.Owner.Id == from {
					filteredWorkflows = append(filteredWorkflows, workflow)
				}
			}
			reassignSummary.Workflows = filteredWorkflows
			reassignSummary.ObjectCounts["workflow"] = len(filteredWorkflows)
		}

		return summaryMsg(&reassignSummary)
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		default:
			return m, nil
		}
	case errMsg:
		m.err = msg
		return m, tea.Quit
	case summaryMsg:
		m.done = true
		m.reassignResult = msg
		return m, tea.Quit
	case reassignDoneMsg:
		m.done = true
		m.reassigning = false
		return m, tea.Quit
	case statusMsg:
		m.statusText = string(msg)
		return m, nil
	case int:
		// Continue to next reassignment step
		apiClient, err := config.InitAPIClient(true)
		if err != nil {
			return m, func() tea.Msg { return errMsg(err) }
		}
		return m, nextReassignmentStepCmd(apiClient.V2024, *m.reassignResult, msg)
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m model) View() string {
	if m.err != nil {
		return m.err.Error()
	}
	if m.done {
		return ""
	}
	var action string

	if m.reassigning {
		action = m.statusText
		if action == "" {
			action = "Beginning reassignment"
		}
	} else {
		action = "Gathering objects to reassign"
	}

	str := fmt.Sprintf("\n\n   %s %s...press q to quit\n\n", m.spinner.View(), action)
	if m.quitting {
		return str + "\n"
	}
	return str
}

func nextReassignmentStepCmd(apiClient *api_v2024.APIClient, summary ReassignSummary, index int) tea.Cmd {
	steps := []struct {
		name      string
		shouldRun bool
		run       func() error
	}{
		{"Reassigning sources", len(summary.Sources) > 0, func() error {
			return reassignSources(apiClient, summary.From, summary.To, summary.Sources)
		}},
		{"Reassigning roles", len(summary.Roles) > 0, func() error {
			return reassignRoles(apiClient, summary.From, summary.To, summary.Roles)
		}},
		{"Reassigning access profiles", len(summary.AccessProfiles) > 0, func() error {
			return reassignAccessProfiles(apiClient, summary.From, summary.To, summary.AccessProfiles)
		}},
		{"Reassigning entitlements", len(summary.Entitlements) > 0, func() error {
			return reassignEntitlements(apiClient, summary.From, summary.To, summary.Entitlements)
		}},
		{"Reassigning identity profiles", len(summary.IdentityProfiles) > 0, func() error {
			return reassignIdentityProfiles(apiClient, summary.From, summary.To, summary.IdentityProfiles)
		}},
		{"Reassigning governance groups", len(summary.GovernanceGroups) > 0, func() error {
			return reassignGovernanceGroups(apiClient, summary.From, summary.To, summary.GovernanceGroups)
		}},
		{"Reassigning workflows", len(summary.Workflows) > 0, func() error {
			return reassignWorkflows(apiClient, summary.From, summary.To, summary.Workflows)
		}},
	}

	if index >= len(steps) {
		return func() tea.Msg { return reassignDoneMsg{} }
	}

	step := steps[index]
	if step.shouldRun {
		return tea.Batch(
			func() tea.Msg { return statusMsg(step.name) },
			func() tea.Msg {
				if err := step.run(); err != nil {
					return errMsg(err)
				}
				return index + 1
			},
		)
	}

	return func() tea.Msg {
		return index + 1
	}
}

func reassignSources(apiClient *api_v2024.APIClient, from Identity, to Identity, sources []api_v2024.Source) error {
	if len(sources) > 0 {
		for _, source := range sources {

			newOwnerId := api_v2024.UpdateMultiHostSourcesRequestInnerValue{String: &to.ID}
			patchArray := []api_v2024.JsonPatchOperation{{Op: "replace", Path: "/owner/id", Value: &newOwnerId}}
			_, _, err := apiClient.SourcesAPI.UpdateSource(context.TODO(), *source.Id).JsonPatchOperation(patchArray).Execute()

			if err != nil {
				fmt.Print("Error updating role owner: ", err)
			}
		}
	}
	return nil
}

func reassignRoles(apiClient *api_v2024.APIClient, from Identity, to Identity, roles []api_v2024.Role) error {
	if len(roles) > 0 {
		for _, role := range roles {
			newOwnerId := api_v2024.UpdateMultiHostSourcesRequestInnerValue{String: &to.ID}
			patchArray := []api_v2024.JsonPatchOperation{{Op: "replace", Path: "/owner/id", Value: &newOwnerId}}
			_, _, err := apiClient.RolesAPI.PatchRole(context.TODO(), *role.Id).JsonPatchOperation(patchArray).Execute()

			if err != nil {
				fmt.Print("Error updating role owner: ", err)
			}
		}
	}
	return nil

}

func reassignAccessProfiles(apiClient *api_v2024.APIClient, from Identity, to Identity, accessProfiles []api_v2024.AccessProfile) error {
	if len(accessProfiles) > 0 {
		for _, accessProfile := range accessProfiles {
			newOwnerId := api_v2024.UpdateMultiHostSourcesRequestInnerValue{String: &to.ID}
			patchArray := []api_v2024.JsonPatchOperation{{Op: "replace", Path: "/owner/id", Value: &newOwnerId}}
			_, _, err := apiClient.AccessProfilesAPI.PatchAccessProfile(context.TODO(), *accessProfile.Id).JsonPatchOperation(patchArray).Execute()

			if err != nil {
				fmt.Print("Error updating access profile owner: ", err)

			}
		}
	}

	return nil
}

func reassignEntitlements(apiClient *api_v2024.APIClient, from Identity, to Identity, entitlements []api_v2024.Entitlement) error {
	if len(entitlements) > 0 {
		for _, entitlement := range entitlements {

			newOwnerId := api_v2024.UpdateMultiHostSourcesRequestInnerValue{String: &to.ID}
			newOwnerName := api_v2024.UpdateMultiHostSourcesRequestInnerValue{String: &to.Name}
			patchArray := []api_v2024.JsonPatchOperation{{Op: "replace", Path: "/owner/id", Value: &newOwnerId}, {Op: "replace", Path: "/owner/name", Value: &newOwnerName}}
			_, _, err := apiClient.EntitlementsAPI.PatchEntitlement(context.TODO(), *entitlement.Id).JsonPatchOperation(patchArray).Execute()

			if err != nil {
				fmt.Print("Error updating entitlement owner: ", err)
			}
		}
	}

	return nil
}

func reassignIdentityProfiles(apiClient *api_v2024.APIClient, from Identity, to Identity, identityProfiles []api_v2024.IdentityProfile) error {
	if len(identityProfiles) > 0 {
		for _, identityProfile := range identityProfiles {
			newOwnerId := api_v2024.UpdateMultiHostSourcesRequestInnerValue{String: &to.ID}
			patchArray := []api_v2024.JsonPatchOperation{{Op: "replace", Path: "/owner/id", Value: &newOwnerId}}
			_, _, err := apiClient.IdentityProfilesAPI.UpdateIdentityProfile(context.TODO(), *identityProfile.Id).JsonPatchOperation(patchArray).Execute()

			if err != nil {
				log.Debug("Error updating identity profile owner: ", err)
			}
		}
	}

	return nil
}

func reassignGovernanceGroups(apiClient *api_v2024.APIClient, from Identity, to Identity, governanceGroups []api_v2024.WorkgroupDto) error {
	if len(governanceGroups) > 0 {
		for _, governanceGroup := range governanceGroups {
			newOwnerId := api_v2024.UpdateMultiHostSourcesRequestInnerValue{String: &to.ID}
			patchArray := []api_v2024.JsonPatchOperation{{Op: "replace", Path: "/owner/id", Value: &newOwnerId}}
			_, _, err := apiClient.GovernanceGroupsAPI.PatchWorkgroup(context.TODO(), *governanceGroup.Id).JsonPatchOperation(patchArray).Execute()

			if err != nil {
				fmt.Print("Error updating governance group owner: ", err)
			}
		}
	}

	return nil
}

func reassignWorkflows(apiClient *api_v2024.APIClient, from Identity, to Identity, workflows []api_v2024.Workflow) error {
	if len(workflows) > 0 {
		for _, workflow := range workflows {

			patchObject := map[string]interface{}{
				"id":   to.ID,
				"type": "IDENTITY",
			}

			newOwner := api_v2024.UpdateMultiHostSourcesRequestInnerValue{MapmapOfStringAny: &patchObject}
			patchArray := []api_v2024.JsonPatchOperation{{Op: "replace", Path: "/owner", Value: &newOwner}}
			_, _, err := apiClient.WorkflowsAPI.PatchWorkflow(context.TODO(), *workflow.Id).JsonPatchOperation(patchArray).Execute()

			if err != nil {
				fmt.Print("Error updating workflow owner: ", err)
			}
		}
	}

	return nil
}
