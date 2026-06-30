package ui_plugins

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/spf13/cobra"
)

const pluginInstancesEndpoint = "/v2026/ui-plugins"

func newCreateCommand() *cobra.Command {
	var private bool
	var restrictToUsers []string
	var dryRun bool
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a UI plugin instance in the current tenant",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := createConfig{
				private:         private,
				restrictToUsers: restrictToUsers,
				dryRun:          dryRun,
				jsonOutput:      jsonOutput,
			}

			// The client is only needed on the non-dry-run path; a dry run never
			// touches config or the network.
			var spClient client.Client
			if !dryRun {
				if err := config.InitConfig(); err != nil {
					return err
				}
				cliConfig, err := config.GetConfig()
				if err != nil {
					return err
				}
				spClient = client.NewSpClient(cliConfig)
			}

			return runCreate(context.Background(), spClient, manifestFileName, cmd.OutOrStdout(), config.GetCurrentIdentityID, opts)
		},
	}

	cmd.Flags().BoolVar(&private, "private", false, "Restrict the plugin to the current user on every slot")
	cmd.Flags().StringSliceVar(&restrictToUsers, "restrict-to-users", nil, "Restrict the plugin to the given user identity GUIDs on every slot (comma-separated)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Validate and print the payload that would be sent without calling the backend")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print the raw UMS response on success")

	return cmd
}

// createConfig captures the create command's flag values.
type createConfig struct {
	private         bool
	restrictToUsers []string
	dryRun          bool
	jsonOutput      bool
}

// runCreate loads and validates the workspace manifest at manifestPath, applies
// visibility overrides, then either prints the payload (dry run) or POSTs it via
// the provided client and renders the result. The client is only used on the
// non-dry-run path, so callers may pass nil for a dry run. currentUser resolves
// the GUID for --private and is injected for testability.
func runCreate(ctx context.Context, c client.Client, manifestPath string, out io.Writer, currentUser func() (string, error), opts createConfig) error {
	cfg, err := loadAndValidateWorkspaceManifest(manifestPath)
	if err != nil {
		return err
	}

	users, err := resolveRestrictToUsers(opts.private, opts.restrictToUsers, currentUser)
	if err != nil {
		return err
	}
	if len(users) > 0 {
		applyVisibilityOverride(&cfg.Manifest, users)
	}

	payload, err := json.MarshalIndent(cfg.Manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode plugin manifest: %w", err)
	}

	if opts.dryRun {
		_, _ = fmt.Fprintln(out, string(payload))
		return nil
	}

	headers := map[string]string{"Accept": "application/json"}
	resp, err := c.Post(ctx, pluginInstancesEndpoint, "application/json", bytes.NewReader(payload), headers)
	if err != nil {
		return fmt.Errorf("failed to create plugin instance: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return mapUMSCreateError(resp.StatusCode, body, cfg.Manifest.Alias)
	}

	if opts.jsonOutput {
		_, _ = fmt.Fprintln(out, string(body))
		return nil
	}

	return renderCreateSuccess(out, body, cfg.Manifest.Alias)
}

// resolveRestrictToUsers builds the de-duplicated, order-preserving union of the
// explicitly provided user GUIDs and, when private is set, the current user's
// identity GUID. currentUser is injected so the resolution path is testable.
func resolveRestrictToUsers(private bool, flagUsers []string, currentUser func() (string, error)) ([]string, error) {
	seen := make(map[string]struct{})
	var users []string

	add := func(user string) {
		user = strings.TrimSpace(user)
		if user == "" {
			return
		}
		if _, ok := seen[user]; ok {
			return
		}
		seen[user] = struct{}{}
		users = append(users, user)
	}

	for _, user := range flagUsers {
		add(user)
	}

	if private {
		id, err := currentUser()
		if err != nil {
			return nil, err
		}
		add(id)
	}

	return users, nil
}

// applyVisibilityOverride sets restrictToUsers on every slot in the manifest,
// giving each slot its own copy of the user list.
func applyVisibilityOverride(manifest *uiPluginManifest, users []string) {
	for i := range manifest.Slots {
		slotUsers := make([]string, len(users))
		copy(slotUsers, users)
		manifest.Slots[i].RestrictToUsers = slotUsers
	}
}

// mapUMSCreateError translates a non-2xx UMS response into an actionable error,
// surfacing the message returned by the backend.
func mapUMSCreateError(status int, body []byte, alias string) error {
	message := umsErrorMessage(body)

	switch status {
	case http.StatusBadRequest:
		return fmt.Errorf("invalid request creating plugin instance: %s", message)
	case http.StatusForbidden:
		return fmt.Errorf("not authorized to create UI plugins (requires the idn:plugins-ui:create right): %s", message)
	case http.StatusNotFound:
		return fmt.Errorf("the UI plugins feature is not enabled for this tenant: %s", message)
	case http.StatusConflict:
		return fmt.Errorf("a plugin instance with alias %q already exists for this tenant: %s", alias, message)
	default:
		return fmt.Errorf("failed to create plugin instance (status %d): %s", status, message)
	}
}

// umsErrorMessage extracts a human-readable message from a UMS error body,
// which follows the NestJS shape {"statusCode":N,"message":string|[]string,"error":string}.
// It falls back to the raw body when the shape is unrecognized.
func umsErrorMessage(body []byte) string {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return "no response body"
	}

	var parsed struct {
		Message json.RawMessage `json:"message"`
		Error   string          `json:"error"`
	}
	if err := json.Unmarshal(body, &parsed); err == nil {
		if len(parsed.Message) > 0 {
			var single string
			if json.Unmarshal(parsed.Message, &single) == nil && strings.TrimSpace(single) != "" {
				return single
			}
			var many []string
			if json.Unmarshal(parsed.Message, &many) == nil && len(many) > 0 {
				return strings.Join(many, "; ")
			}
		}
		if parsed.Error != "" {
			return parsed.Error
		}
	}

	return trimmed
}

// renderCreateSuccess prints a human-readable confirmation of the created instance.
func renderCreateSuccess(w io.Writer, body []byte, fallbackAlias string) error {
	var created struct {
		PluginInstanceID string `json:"pluginInstanceId"`
		Alias            string `json:"alias"`
	}
	_ = json.Unmarshal(body, &created)

	alias := created.Alias
	if alias == "" {
		alias = fallbackAlias
	}

	if created.PluginInstanceID != "" {
		_, _ = fmt.Fprintf(w, "Created plugin instance %s (alias: %s)\n", created.PluginInstanceID, alias)
	} else {
		_, _ = fmt.Fprintf(w, "Created plugin instance (alias: %s)\n", alias)
	}

	return nil
}
