package compliance

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	sailpoint "github.com/sailpoint-oss/golang-sdk/v2"
	beta "github.com/sailpoint-oss/golang-sdk/v2/api_beta"
	api_v2024 "github.com/sailpoint-oss/golang-sdk/v2/api_v2024"
	v3 "github.com/sailpoint-oss/golang-sdk/v2/api_v3"
	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

//go:embed collect.md
var collectHelp string

type collectorContext struct {
	apiClient *sailpoint.APIClient
	rawClient client.Client
	period    int
}

func newCollectCommand() *cobra.Command {
	help := util.ParseHelp(collectHelp)

	var outputFile string
	var periodDays int
	var pretty bool

	cmd := &cobra.Command{
		Use:     "collect",
		Short:   "Collect tenant evidence into a compliance bundle",
		Long:    help.Long,
		Example: help.Example,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if periodDays < 1 {
				return fmt.Errorf("period must be greater than 0")
			}

			if err := config.InitConfig(); err != nil {
				return err
			}

			apiClient, err := config.InitAPIClient(false)
			if err != nil {
				return err
			}

			cfg, err := config.GetConfig()
			if err != nil {
				return err
			}

			metadataTenant := config.GetTenantUrl()
			if strings.TrimSpace(metadataTenant) == "" {
				metadataTenant = config.GetActiveEnvironment()
			}

			bundle := EvidenceBundle{
				Metadata: Metadata{
					SchemaVersion:  "1.0.0",
					GeneratedAt:    time.Now().UTC(),
					SailCLIVersion: sailVersion(cmd),
					PeriodDays:     periodDays,
					Tenant:         metadataTenant,
				},
				Data: EvidenceData{
					Events: EventSummary{PeriodDays: periodDays},
				},
				Summary: CollectionSummary{
					Errors: []string{},
				},
			}

			collectors := []struct {
				name string
				run  func(context.Context, *collectorContext, *EvidenceBundle) error
			}{
				{name: "auth_org_config", run: collectAuthOrgConfig},
				{name: "password_policies", run: collectPasswordPolicies},
				{name: "sod_policies", run: collectSODPolicies},
				{name: "certifications", run: collectCertifications},
				{name: "identities", run: collectIdentities},
				{name: "roles", run: collectRoles},
				{name: "access_profiles", run: collectAccessProfiles},
				{name: "sources", run: collectSources},
				{name: "lifecycle_states", run: collectLifecycleStates},
				{name: "workflows", run: collectWorkflows},
				{name: "governance_groups", run: collectGovernanceGroups},
				{name: "provisioning_events", run: collectProvisioningEvents},
				{name: "password_events", run: collectPasswordEvents},
			}

			bundle.Summary.TotalCollectors = len(collectors)

			ctx := context.Background()
			state := &collectorContext{
				apiClient: apiClient,
				rawClient: client.NewSpClient(cfg),
				period:    periodDays,
			}

			for _, collector := range collectors {
				log.Info("Running compliance collector", "collector", collector.name)
				err := collector.run(ctx, state, &bundle)
				if err != nil {
					bundle.Summary.Failed++
					message := fmt.Sprintf("%s: %v", collector.name, err)
					bundle.Summary.Errors = append(bundle.Summary.Errors, message)
					log.Error("Collector failed", "collector", collector.name, "error", err)
					continue
				}
				bundle.Summary.Succeeded++
			}

			if len(bundle.Summary.Errors) == 0 {
				bundle.Summary.Errors = nil
			}

			if err := writeJSONOutput(outputFile, bundle, pretty); err != nil {
				return err
			}

			log.Info("Compliance evidence bundle written", "output", outputFile)

			if bundle.Summary.Failed > 0 {
				return fmt.Errorf("collection completed with %d failed collectors", bundle.Summary.Failed)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFile, "output", "o", "evidence.json", "Output file path for evidence bundle")
	cmd.Flags().IntVarP(&periodDays, "period", "p", 90, "Lookback period in days for event collectors")
	cmd.Flags().BoolVar(&pretty, "pretty", false, "Pretty-print output JSON")

	return cmd
}

func collectAuthOrgConfig(ctx context.Context, state *collectorContext, bundle *EvidenceBundle) error {
	resp, err := state.rawClient.Get(ctx, "/v3/auth-org", map[string]string{"Accept": "application/json"})
	if err == nil {
		defer resp.Body.Close()
		body, readErr := io.ReadAll(resp.Body)
		if readErr == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			bundle.Data.AuthOrgConfig = json.RawMessage(body)
			return nil
		}
	}

	objects, exportErr := exportSPConfigObjects(ctx, state.apiClient, "AUTH_ORG")
	if exportErr != nil {
		if err != nil {
			return fmt.Errorf("raw auth-org request failed (%v); fallback export failed (%w)", err, exportErr)
		}
		return fmt.Errorf("fallback export failed: %w", exportErr)
	}
	if len(objects) == 0 {
		return fmt.Errorf("auth org export returned no objects")
	}
	if len(objects) > 1 {
		log.Warn("AUTH_ORG export returned multiple objects, selecting first", "count", len(objects))
	}

	raw, err := marshalRawMessage(objects[0])
	if err != nil {
		return err
	}
	bundle.Data.AuthOrgConfig = raw
	return nil
}

func collectPasswordPolicies(ctx context.Context, state *collectorContext, bundle *EvidenceBundle) error {
	results, _, err := sailpoint.PaginateWithDefaults[beta.PasswordPolicyV3Dto](state.apiClient.Beta.PasswordPoliciesAPI.ListPasswordPolicies(ctx))
	if err != nil {
		return err
	}
	raw, err := marshalRawMessage(results)
	if err != nil {
		return err
	}
	bundle.Data.PasswordPolicies = raw
	return nil
}

func collectSODPolicies(ctx context.Context, state *collectorContext, bundle *EvidenceBundle) error {
	results, _, err := sailpoint.PaginateWithDefaults[beta.SodPolicy](state.apiClient.Beta.SODPoliciesAPI.ListSodPolicies(ctx))
	if err != nil {
		return err
	}
	raw, err := marshalRawMessage(results)
	if err != nil {
		return err
	}
	bundle.Data.SODPolicies = raw
	return nil
}

func collectCertifications(ctx context.Context, state *collectorContext, bundle *EvidenceBundle) error {
	results, _, err := sailpoint.PaginateWithDefaults[beta.CertificationDto](state.apiClient.Beta.CertificationsAPI.ListCertifications(ctx))
	if err != nil {
		return err
	}
	raw, err := marshalRawMessage(results)
	if err != nil {
		return err
	}
	bundle.Data.Certifications = raw
	return nil
}

func collectIdentities(ctx context.Context, state *collectorContext, bundle *EvidenceBundle) error {
	results, err := searchAll(ctx, state.apiClient, "*", v3.INDEX_IDENTITIES)
	if err != nil {
		return err
	}
	raw, err := marshalRawMessage(results)
	if err != nil {
		return err
	}
	bundle.Data.Identities = raw
	return nil
}

func collectRoles(ctx context.Context, state *collectorContext, bundle *EvidenceBundle) error {
	results, err := searchAll(ctx, state.apiClient, "*", v3.INDEX_ROLES)
	if err != nil {
		return err
	}
	raw, err := marshalRawMessage(results)
	if err != nil {
		return err
	}
	bundle.Data.Roles = raw
	return nil
}

func collectAccessProfiles(ctx context.Context, state *collectorContext, bundle *EvidenceBundle) error {
	results, err := searchAll(ctx, state.apiClient, "*", v3.INDEX_ACCESSPROFILES)
	if err != nil {
		return err
	}
	raw, err := marshalRawMessage(results)
	if err != nil {
		return err
	}
	bundle.Data.AccessProfiles = raw
	return nil
}

func collectSources(ctx context.Context, state *collectorContext, bundle *EvidenceBundle) error {
	results, _, err := sailpoint.PaginateWithDefaults[api_v2024.Source](state.apiClient.V2024.SourcesAPI.ListSources(ctx))
	if err != nil {
		return err
	}
	raw, err := marshalRawMessage(results)
	if err != nil {
		return err
	}
	bundle.Data.Sources = raw
	return nil
}

func collectLifecycleStates(ctx context.Context, state *collectorContext, bundle *EvidenceBundle) error {
	objects, err := exportSPConfigObjects(ctx, state.apiClient, "LIFECYCLE_STATE")
	if err != nil {
		return err
	}
	raw, err := marshalRawMessage(objects)
	if err != nil {
		return err
	}
	bundle.Data.LifecycleStates = raw
	return nil
}

func collectWorkflows(ctx context.Context, state *collectorContext, bundle *EvidenceBundle) error {
	results, _, err := sailpoint.PaginateWithDefaults[beta.Workflow](state.apiClient.Beta.WorkflowsAPI.ListWorkflows(ctx))
	if err != nil {
		return err
	}
	raw, err := marshalRawMessage(results)
	if err != nil {
		return err
	}
	bundle.Data.Workflows = raw
	return nil
}

func collectGovernanceGroups(ctx context.Context, state *collectorContext, bundle *EvidenceBundle) error {
	results, _, err := sailpoint.PaginateWithDefaults[api_v2024.WorkgroupDto](state.apiClient.V2024.GovernanceGroupsAPI.ListWorkgroups(ctx))
	if err != nil {
		return err
	}
	raw, err := marshalRawMessage(results)
	if err != nil {
		return err
	}
	bundle.Data.GovernanceGroups = raw
	return nil
}

func collectProvisioningEvents(ctx context.Context, state *collectorContext, bundle *EvidenceBundle) error {
	query := fmt.Sprintf("(type:provisioning AND created:[now-%dd TO now])", state.period)
	results, err := searchAll(ctx, state.apiClient, query, v3.INDEX_EVENTS)
	if err != nil {
		return err
	}
	bundle.Data.Events.ProvisioningCount = len(results)
	return nil
}

func collectPasswordEvents(ctx context.Context, state *collectorContext, bundle *EvidenceBundle) error {
	query := fmt.Sprintf("(type:PASSWORD_ACTION AND created:[now-%dd TO now])", state.period)
	results, err := searchAll(ctx, state.apiClient, query, v3.INDEX_EVENTS)
	if err != nil {
		return err
	}
	bundle.Data.Events.PasswordCount = len(results)
	return nil
}

func exportSPConfigObjects(ctx context.Context, apiClient *sailpoint.APIClient, includeType string) ([]map[string]interface{}, error) {
	description := fmt.Sprintf("compliance collect %s", includeType)
	job, _, err := apiClient.Beta.SPConfigAPI.ExportSpConfig(ctx).ExportPayload(beta.ExportPayload{
		Description:  &description,
		IncludeTypes: []string{includeType},
	}).Execute()
	if err != nil {
		return nil, err
	}

	for attempt := 0; attempt < 90; attempt++ {
		status, _, err := apiClient.Beta.SPConfigAPI.GetSpConfigExportStatus(ctx, job.JobId).Execute()
		if err != nil {
			return nil, err
		}

		switch status.Status {
		case "NOT_STARTED", "IN_PROGRESS":
			time.Sleep(2 * time.Second)
			continue
		case "COMPLETE":
			exported, _, err := apiClient.Beta.SPConfigAPI.GetSpConfigExport(ctx, job.JobId).Execute()
			if err != nil {
				return nil, err
			}
			objects := make([]map[string]interface{}, 0, len(exported.Objects))
			for _, obj := range exported.Objects {
				if obj.Self != nil && obj.Self.Type != nil && *obj.Self.Type != includeType {
					continue
				}
				if obj.Object != nil {
					objects = append(objects, obj.Object)
				}
			}
			return objects, nil
		case "FAILED":
			return nil, fmt.Errorf("spconfig export failed for %s", includeType)
		case "CANCELLED":
			return nil, fmt.Errorf("spconfig export cancelled for %s", includeType)
		default:
			return nil, fmt.Errorf("unexpected spconfig export status for %s: %s", includeType, status.Status)
		}
	}

	return nil, fmt.Errorf("timed out waiting for spconfig export for %s", includeType)
}

func searchAll(ctx context.Context, apiClient *sailpoint.APIClient, query string, index v3.Index) ([]map[string]interface{}, error) {
	search := v3.NewSearch()
	search.SetIndices([]v3.Index{index})
	queryObj := v3.NewQuery()
	queryObj.SetQuery(query)
	search.SetQuery(*queryObj)

	const limit int32 = 250
	var offset int32
	results := make([]map[string]interface{}, 0)

	for {
		page, resp, err := apiClient.V3.SearchAPI.SearchPost(ctx).Search(*search).Limit(limit).Offset(offset).Execute()
		if err != nil {
			if resp != nil {
				return nil, fmt.Errorf("search failed for index %s query %q: %s: %w", index, query, resp.Status, err)
			}
			return nil, fmt.Errorf("search failed for index %s query %q: %w", index, query, err)
		}

		results = append(results, page...)
		if len(page) < int(limit) {
			break
		}
		offset += limit
	}

	return results, nil
}

func marshalRawMessage(value interface{}) (json.RawMessage, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(payload), nil
}

func writeJSONOutput(outputPath string, value interface{}, pretty bool) error {
	dir := filepath.Dir(outputPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	var payload []byte
	var err error
	if pretty {
		payload, err = json.MarshalIndent(value, "", "  ")
	} else {
		payload, err = json.Marshal(value)
	}
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, payload, 0o644)
}

func sailVersion(cmd *cobra.Command) string {
	root := cmd.Root()
	if root == nil {
		return "unknown"
	}
	if strings.TrimSpace(root.Version) == "" {
		return "unknown"
	}
	return root.Version
}
