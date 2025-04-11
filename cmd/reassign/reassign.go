// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package reassign

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
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

			p := tea.NewProgram(initialModel(from, to, objectTypes, dryRun))
			finalModel, err := p.Run()
			if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}

			if m, ok := finalModel.(model); ok && m.reassignResult != nil {
				p.Quit()
				printSummary(*m.reassignResult)

				// If this was not a dry run proceed with the reassignment flow
				if !m.reassignResult.DryRun {
					fmt.Println("Would you like to save the full report to a file (y/n): ")
					var response string
					_, err := fmt.Scanln(&response)
					if err != nil {
						fmt.Println("Failed to read input:", err)
						return err
					}

					response = strings.ToLower(strings.TrimSpace(response))
					if response == "y" {
						fmt.Print("Enter the file name (without extension)(default: reassign_report): ")
						var fileName string
						_, err := fmt.Scanln(&fileName)

						if err != nil && err.Error() != "unexpected newline" {
							fmt.Println("Failed to read input:", err)
							return nil
						}

						fileName = strings.TrimSpace(fileName)
						if fileName == "" {
							fileName = "reassign_report"
						}
						// Save the report to a file
						reportPath := fmt.Sprintf("%s.json", fileName)

						err = writeReport(*m.reassignResult, reportPath)

						if err != nil {
							fmt.Println("Failed to write report:", err)
						} else {
							fmt.Printf("Report saved to %s\n", reportPath)
						}

						return nil
					}

					fmt.Printf("Would you like to proceed with reassigning these objects from '%s' to '%s': ", m.reassignResult.From.Name, m.reassignResult.To.Name)
					var reassignResponse string
					_, err = fmt.Scanln(&reassignResponse)
					if err != nil {
						fmt.Println("Failed to read input:", err)
						return err
					}

					response = strings.ToLower(strings.TrimSpace(reassignResponse))

					if response == "y" {
						m := initialModel(from, to, objectTypes, dryRun)
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

					// If this was a dry run, just print the summary and allow the user the option to save the report
					fmt.Println("Would you like to save the full report to a file (y/n): ")
					var response string
					_, err := fmt.Scanln(&response)
					if err != nil {
						fmt.Println("Failed to read input:", err)
						return err
					}

					response = strings.ToLower(strings.TrimSpace(response))
					if response == "y" {
						fmt.Print("Enter the file name (without extension)(default: reassign_report): ")
						var fileName string
						_, err := fmt.Scanln(&fileName)

						if err != nil && err.Error() != "unexpected newline" {
							fmt.Println("Failed to read input:", err)
							return nil
						}

						fileName = strings.TrimSpace(fileName)
						if fileName == "" {
							fileName = "reassign_report"
						}
						// Save the report to a file
						reportPath := fmt.Sprintf("%s.json", fileName)

						err = writeReport(*m.reassignResult, reportPath)

						if err != nil {
							fmt.Println("Failed to write report:", err)
						} else {
							fmt.Printf("Report saved to %s\n", reportPath)
						}

						return nil
					} else {
						fmt.Println("Aborted reassignment.")
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&from, "from", "f", "", "The identity to reassign from")
	cmd.Flags().StringVarP(&to, "to", "t", "", "The identity to reassign to")
	cmd.Flags().BoolVarP(&force, "force", "F", false, "Bypass confirmation prompt")
	cmd.Flags().StringVarP(&objectTypes, "object-types", "o", "", "Comma-separated list of object types to reassign, defaults to all")
	cmd.Flags().StringVarP(&objectId, "object-id", "i", "", "The object id to reassign")
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Show the objects that would be reassigned without actually reassigning them")

	return cmd

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
	err            error
	done           bool
	reassigning    bool
	statusText     string
	reassignResult *ReassignSummary
}

func initialModel(from string, to string, objectTypes string, dryRun bool) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return model{spinner: s, from: from, to: to, objectTypes: objectTypes, dryRun: dryRun}
}

func (m model) Init() tea.Cmd {
	if m.reassigning && m.reassignResult != nil {
		apiClient, err := config.InitAPIClient(true)
		if err != nil {
			return func() tea.Msg { return errMsg(err) }
		}
		return tea.Batch(m.spinner.Tick, nextReassignmentStepCmd(apiClient.V2024, *m.reassignResult, 0))
	}
	return tea.Batch(m.spinner.Tick, fetchReassignSummaryCmd(m.from, m.to, m.objectTypes, m.dryRun))
}

func fetchReassignSummaryCmd(from string, to string, objectTypes string, dryRun bool) tea.Cmd {
	return func() tea.Msg {
		// your logic here (init API, gather data, etc)
		// return errMsg(err) on error or summaryMsg(result)
		var objectsToReassign []string
		var reassignIdentities []api_v2024.Identity
		var sources []api_v2024.Source
		var roles []api_v2024.Role
		var accessProfiles []api_v2024.AccessProfile
		var filteredIdentityProfiles []api_v2024.IdentityProfile
		var entitlements []api_v2024.Entitlement
		var filteredGovernanceGroups []api_v2024.WorkgroupDto
		var filteredWorkflows []api_v2024.Workflow

		apiClient, err := config.InitAPIClient(true)

		if err != nil {
			return err
		}

		if from != "" && to != "" {
			filters := fmt.Sprintf("id in (\"%s\",\"%s\")", from, to)
			identities, _, err := apiClient.V2024.IdentitiesAPI.ListIdentities(context.TODO()).Filters(filters).Execute()
			if err != nil {
				return err
			}
			if len(identities) != 2 {
				return fmt.Errorf("expected 2 identities, got %d", len(identities))
			}
			reassignIdentities = identities
		}

		if objectTypes != "" {
			err := validateObjectTypes(objectTypes)
			if err != nil {
				log.Error(err)
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
			sources, _, err = apiClient.V2024.SourcesAPI.ListSources(context.TODO()).Filters(filters).Execute()
			if err != nil {
				return err
			}

			reassignSummary.Sources = sources
			reassignSummary.ObjectCounts["source"] = len(sources)

		}

		if contains(objectsToReassign, "role") {
			log.Debug("Gathering roles to reassign")
			filters := fmt.Sprintf("owner.id eq \"%s\"", from)
			roles, _, err = apiClient.V2024.RolesAPI.ListRoles(context.TODO()).Filters(filters).Execute()
			if err != nil {
				return err
			}

			reassignSummary.Roles = roles
			reassignSummary.ObjectCounts["role"] = len(roles)
		}

		if contains(objectsToReassign, "access-profile") {
			log.Debug("Gathering access profiles to reassign")
			filters := fmt.Sprintf("owner.id eq \"%s\"", from)
			accessProfiles, _, err = apiClient.V2024.AccessProfilesAPI.ListAccessProfiles(context.TODO()).Filters(filters).Execute()
			if err != nil {
				return err
			}
			reassignSummary.AccessProfiles = accessProfiles
			reassignSummary.ObjectCounts["access-profile"] = len(accessProfiles)
		}

		if contains(objectsToReassign, "entitlement") {
			log.Debug("Gathering entitlements to reassign")
			filters := fmt.Sprintf("owner.id eq \"%s\"", from)
			entitlements, _, err = apiClient.V2024.EntitlementsAPI.ListEntitlements(context.TODO()).Filters(filters).Execute()
			if err != nil {
				return err
			}
			reassignSummary.Entitlements = entitlements
			reassignSummary.ObjectCounts["entitlement"] = len(entitlements)
		}

		if contains(objectsToReassign, "identity-profile") {
			log.Debug("Gathering identity profiles to reassign")
			identityProfiles, _, err := apiClient.V2024.IdentityProfilesAPI.ListIdentityProfiles(context.TODO()).Execute()
			if err != nil {
				return err
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
			governanceGroups, _, err := apiClient.V2024.GovernanceGroupsAPI.ListWorkgroups(context.TODO()).Execute()
			if err != nil {
				return err
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

		if contains(objectsToReassign, "workflow") {
			log.Debug("Gathering workflows to reassign")
			workflows, _, err := apiClient.V2024.WorkflowsAPI.ListWorkflows(context.TODO()).Execute()
			if err != nil {
				return err
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
			action = "Reassigning objects"
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
			return reassignTest(apiClient, summary.From, summary.To, summary.Sources)
		}},
		{"Reassigning roles", len(summary.Roles) > 0, func() error {
			return reassignTest(apiClient, summary.From, summary.To, summary.Roles)
		}},
		{"Reassigning access profiles", len(summary.AccessProfiles) > 0, func() error {
			return reassignTest(apiClient, summary.From, summary.To, summary.AccessProfiles)
		}},
		{"Reassigning entitlements", len(summary.Entitlements) > 0, func() error {
			return reassignTest(apiClient, summary.From, summary.To, summary.Entitlements)
		}},
		{"Reassigning identity profiles", len(summary.IdentityProfiles) > 0, func() error {
			return reassignTest(apiClient, summary.From, summary.To, summary.IdentityProfiles)
		}},
		{"Reassigning governance groups", len(summary.GovernanceGroups) > 0, func() error {
			return reassignTest(apiClient, summary.From, summary.To, summary.GovernanceGroups)
		}},
		{"Reassigning workflows", len(summary.Workflows) > 0, func() error {
			return reassignTest(apiClient, summary.From, summary.To, summary.Workflows)
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

func reassignTest[T any](apiClient *api_v2024.APIClient, from Identity, to Identity, sources []T) error {
	time.Sleep(time.Second * 3)
	return nil
}

func reassignSources(apiClient *api_v2024.APIClient, from Identity, to Identity, sources []api_v2024.Source) error {
	if len(sources) > 0 {
		for _, source := range sources {

			newOwnerId := api_v2024.UpdateMultiHostSourcesRequestInnerValue{String: &to.ID}
			patchArray := []api_v2024.JsonPatchOperation{{Op: "replace", Path: "/owner/id", Value: &newOwnerId}}
			_, _, err := apiClient.SourcesAPI.UpdateSource(context.TODO(), *source.Id).JsonPatchOperation(patchArray).Execute()

			if err != nil {
				log.Debug("Error updating source: ", err)
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
				log.Debug("Error updating role: ", err)
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
				log.Debug("Error updating access profile: ", err)
			}
		}
	}

	return nil
}

func reassignEntitlements(apiClient *api_v2024.APIClient, from Identity, to Identity, entitlements []api_v2024.Entitlement) error {
	if len(entitlements) > 0 {
		entitlementIds := make([]string, len(entitlements))
		for i, item := range entitlements {
			entitlementIds[i] = *item.Id
		}
		patch := api_v2024.EntitlementBulkUpdateRequest{
			EntitlementIds: entitlementIds,
			JsonPatch:      []api_v2024.JsonPatchOperation{},
		}

		_, err := apiClient.EntitlementsAPI.UpdateEntitlementsInBulk(context.TODO()).EntitlementBulkUpdateRequest(patch).Execute()

		if err != nil {
			log.Debug("Error updating entitlement: ", err)
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
				log.Debug("Error updating access profile: ", err)
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
				log.Debug("Error updating access profile: ", err)
			}
		}
	}

	return nil
}

func reassignWorkflows(apiClient *api_v2024.APIClient, from Identity, to Identity, workflows []api_v2024.Workflow) error {
	if len(workflows) > 0 {
		for _, workflow := range workflows {
			newOwnerId := api_v2024.UpdateMultiHostSourcesRequestInnerValue{String: &to.ID}
			patchArray := []api_v2024.JsonPatchOperation{{Op: "replace", Path: "/owner/id", Value: &newOwnerId}}
			_, _, err := apiClient.WorkflowsAPI.PatchWorkflow(context.TODO(), *workflow.Id).JsonPatchOperation(patchArray).Execute()

			if err != nil {
				log.Debug("Error updating access profile: ", err)
			}
		}
	}

	return nil
}
