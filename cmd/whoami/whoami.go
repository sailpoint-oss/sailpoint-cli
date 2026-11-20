package whoami

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

//go:embed whoami.md
var whoamiHelp string

func NewWhoamiCommand() *cobra.Command {
	var jsonOutput bool

	help := util.ParseHelp(whoamiHelp)

	cmd := &cobra.Command{
		Use:     "whoami",
		Short:   "Show how the CLI is logged in, including user, tenant and scopes",
		Long:    help.Long,
		Example: help.Example,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			envFlag := cmd.Flags().Lookup("env")
			info, err := config.GetAuthInfo(envFlag != nil && envFlag.Changed)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			if jsonOutput {
				encoded, err := json.MarshalIndent(info, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to encode whoami output: %w", err)
				}
				_, _ = fmt.Fprintln(out, string(encoded))
				return nil
			}

			printInfo(out, info, time.Now())
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print the result as JSON")

	return cmd
}

// printInfo renders info as labelled sections, with warnings first. now is passed
// in so tests can control relative times such as "expires in".
func printInfo(out io.Writer, info config.AuthInfo, now time.Time) {
	if len(info.Warnings) > 0 {
		_, _ = fmt.Fprintln(out, "Warnings")
		for _, warning := range info.Warnings {
			_, _ = fmt.Fprintf(out, "  ! %s\n", warning)
		}
		_, _ = fmt.Fprintln(out)
	}

	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	row := func(label, value, source string) {
		if value == "" {
			value = "-"
		}
		if source != "" {
			value = fmt.Sprintf("%s  (%s)", value, source)
		}
		_, _ = fmt.Fprintf(w, "%s:\t%s\n", label, value)
	}

	_, _ = fmt.Fprintln(w, "Configuration")
	row("  Environment", info.Environment, info.EnvironmentSource)
	row("  Auth type", strings.ToUpper(info.AuthType), info.AuthTypeSource)
	row("  Base URL", info.BaseURL, info.BaseURLSource)
	if info.AuthType == "oauth" {
		row("  Tenant URL", info.TenantURL, info.TenantURLSource)
	}
	if info.AuthType == "pat" {
		row("  Client ID", fromSource(info.ClientIDSource), "")
		row("  Client secret", fromSource(info.ClientSecretSource), "")
	}

	_, _ = fmt.Fprintln(w)
	if len(info.EnvVars) == 0 {
		_, _ = fmt.Fprintln(w, "Environment variables\tnone set")
	} else {
		_, _ = fmt.Fprintln(w, "Environment variables (these override the config file and keychain)")
		for _, v := range info.EnvVars {
			row("  "+v.Name, v.Value, "")
		}
	}

	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Token")
	if !info.LoggedIn {
		row("  Status", "not logged in", "")
		row("  Reason", oneLine(info.TokenError), "")
		w.Flush()
		if info.AuthType == "oauth" && info.TokenError == config.ErrNotLoggedIn.Error() {
			_, _ = fmt.Fprintln(out, "\nRun any command that calls the API, such as `sail search query`, to log in with OAuth.")
		}
		return
	}

	c := info.Claims
	if c.HasUserContext() {
		row("  User", c.UserName, "")
		row("  Identity ID", c.IdentityID, "")
	} else {
		row("  User", "none (client credentials without a user context)", "")
	}
	row("  Client ID", c.ClientID, "")
	row("  Tenant", c.Org, "")
	row("  Pod", c.Pod, "")
	row("  Strong auth", fmt.Sprintf("%t", c.StrongAuth), "")
	if !c.ExpiresAt.IsZero() {
		status := fmt.Sprintf("expires in %s", c.ExpiresAt.Sub(now).Round(time.Second))
		if !c.ExpiresAt.After(now) {
			status = "expired"
		}
		row("  Expires", c.ExpiresAt.Local().Format(time.RFC3339), status)
	}
	row("  Authorities", strings.Join(c.Authorities, ", "), "")
	if c.ScopesEncoded {
		note := "the CLI can't decode them"
		if info.AuthType == "pat" {
			note = "see the PAT's scopes below"
		}
		row("  Scopes", "encoded in the token: "+strings.Join(c.Scopes, " "), note)
	} else {
		row("  Scopes", strings.Join(c.Scopes, " "), "")
	}

	if info.AuthType == "pat" {
		_, _ = fmt.Fprintln(w)
		_, _ = fmt.Fprintln(w, "Personal access token")
		if info.PAT == nil {
			reason := info.PATLookupError
			if reason == "" {
				reason = "not available for tokens without a user context"
			}
			row("  Details", "unavailable", reason)
		} else {
			p := info.PAT
			row("  Name", p.Name, "")
			row("  ID", p.ID, "")
			row("  Owner", p.Owner, "")
			row("  Scopes", strings.Join(p.Scopes, " "), "")
			row("  Created", p.Created.Local().Format("2006-01-02"), age(p.Created, now))
			if p.ExpiresAt != nil {
				row("  Expires", p.ExpiresAt.Local().Format("2006-01-02"), "")
			} else {
				row("  Expires", "never", "")
			}
		}
	}
	w.Flush()
}

// age describes how long ago t was in days, or years and days once it's over a year.
func age(t, now time.Time) string {
	days := int(now.Sub(t).Hours() / 24)
	if days < 365 {
		return fmt.Sprintf("%d days ago", days)
	}
	return fmt.Sprintf("%d years, %d days ago", days/365, days%365)
}

// fromSource describes where a PAT credential comes from, such as "from keychain".
func fromSource(source string) string {
	if source == config.SourceNone {
		return source
	}
	return "from " + source
}

// oneLine collapses a multi-line error message so it fits in one aligned row.
// A line that doesn't end in punctuation gets a period, so the sentences don't run together.
func oneLine(s string) string {
	var parts []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.Join(strings.Fields(line), " ")
		if line == "" {
			continue
		}
		if n := len(parts); n > 0 && !strings.ContainsAny(parts[n-1][len(parts[n-1])-1:], ".!?:") {
			parts[n-1] += "."
		}
		parts = append(parts, line)
	}
	return strings.Join(parts, " ")
}
