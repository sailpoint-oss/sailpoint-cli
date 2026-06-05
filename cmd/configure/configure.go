package configure

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/clierror"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var validKeys = map[string]string{
	"debug":                 "Enable or disable debug logging (true/false)",
	"export-templates-path": "Path to custom SPConfig export templates file",
	"search-templates-path": "Path to custom search templates file",
	"report-templates-path": "Path to custom report templates file",
}

var keyMapping = map[string]string{
	"debug":                 "debug",
	"export-templates-path": "exporttemplatespath",
	"search-templates-path": "searchtemplatespath",
	"report-templates-path": "reporttemplatespath",
}

type globalConfig struct {
	Debug               bool   `json:"debug" yaml:"debug"`
	ActiveEnvironment   string `json:"activeEnvironment" yaml:"activeEnvironment"`
	ExportTemplatesPath string `json:"exportTemplatesPath" yaml:"exportTemplatesPath"`
	SearchTemplatesPath string `json:"searchTemplatesPath" yaml:"searchTemplatesPath"`
	ReportTemplatesPath string `json:"reportTemplatesPath" yaml:"reportTemplatesPath"`
}

func NewConfigureCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config [key] [value]",
		Short: "View or modify global CLI settings",
		Long:  "\nView or modify global CLI settings such as debug mode and template paths.\n\nWith no arguments, displays all settings.\nWith one argument, displays the value of that key.\nWith two arguments, sets the key to the given value.\n",
		Example: `  sail config                                   # show all settings
  sail config debug                             # get debug value
  sail config debug true                        # set debug to true
  sail config export-templates-path /path/to/f  # set a template path`,
		Args: cobra.MaximumNArgs(2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) > 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			keys := make([]string, 0, len(keyMapping))
			for key := range keyMapping {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			return keys, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			w := cmd.OutOrStdout()
			switch len(args) {
			case 0:
				return listConfig(w)
			case 1:
				return getConfig(w, args[0])
			case 2:
				return setConfig(args[0], args[1])
			}
			return nil
		},
	}

	return cmd
}

func listConfig(w io.Writer) error {
	cfg := globalConfig{
		Debug:               config.GetDebug(),
		ActiveEnvironment:   config.GetActiveEnvironment(),
		ExportTemplatesPath: config.GetCustomExportTemplatePath(),
		SearchTemplatesPath: config.GetCustomSearchTemplatePath(),
		ReportTemplatesPath: config.GetCustomReportTemplatePath(),
	}
	if output.IsMachineReadable() {
		return output.WriteStructured(w, cfg)
	}

	fmt.Fprintln(w, "Global CLI Configuration:")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "  %-25s %v\n", "debug", cfg.Debug)
	fmt.Fprintf(w, "  %-25s %s\n", "active-environment", cfg.ActiveEnvironment)
	fmt.Fprintf(w, "  %-25s %s\n", "export-templates-path", cfg.ExportTemplatesPath)
	fmt.Fprintf(w, "  %-25s %s\n", "search-templates-path", cfg.SearchTemplatesPath)
	fmt.Fprintf(w, "  %-25s %s\n", "report-templates-path", cfg.ReportTemplatesPath)
	return nil
}

func getConfig(w io.Writer, key string) error {
	viperKey, ok := keyMapping[key]
	if !ok {
		return unknownKey(w, key)
	}
	value := viper.Get(viperKey)
	if output.IsMachineReadable() {
		return output.WriteStructured(w, map[string]any{key: value})
	}
	fmt.Fprintf(w, "%s = %v\n", key, value)
	return nil
}

func setConfig(key, value string) error {
	switch key {
	case "debug":
		switch strings.ToLower(value) {
		case "true", "enable":
			config.SetDebug(true)
			log.Info("Debug mode enabled")
		case "false", "disable":
			config.SetDebug(false)
			log.Info("Debug mode disabled")
		default:
			return fmt.Errorf("invalid value for debug: %s (use true/false)", value)
		}

	case "export-templates-path":
		config.SetCustomExportTemplatePath(value)
		log.Info("Export templates path updated", "path", value)

	case "search-templates-path":
		config.SetCustomSearchTemplatePath(value)
		log.Info("Search templates path updated", "path", value)

	case "report-templates-path":
		config.SetCustomReportTemplatePath(value)
		log.Info("Report templates path updated", "path", value)

	default:
		return clierror.Usage("unknown config key: " + key)
	}

	return nil
}

func unknownKey(w io.Writer, key string) error {
	fmt.Fprintf(w, "Unknown config key: %s\n\nValid keys:\n", key)
	for k, desc := range validKeys {
		fmt.Fprintf(w, "  %-25s %s\n", k, desc)
	}
	return clierror.Usage("unknown config key: " + key)
}
