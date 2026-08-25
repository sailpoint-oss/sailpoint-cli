package ui_plugins

// uiPluginWorkspaceConfig represents the sp-ui-plugin.json workspace contract.
type uiPluginWorkspaceConfig struct {
	Version  int                  `json:"version"`
	Manifest uiPluginManifest     `json:"manifest"`
	Build    *uiPluginBuildConfig `json:"build,omitempty"`
}

// uiPluginSlot declares a UI extension point the plugin occupies.
type uiPluginSlot struct {
	SlotID               string   `json:"slotId"`
	RequiredCapabilities []string `json:"requiredCapabilities,omitempty"`
	RestrictToUsers      []string `json:"restrictToUsers,omitempty"`
}

// uiPluginManifest is the backend-facing payload section.
type uiPluginManifest struct {
	Alias                   string              `json:"alias"`
	Name                    map[string]string   `json:"name"`
	Description             map[string]string   `json:"description"`
	APIScopes               []string            `json:"apiScopes,omitempty"`
	ContentSecurityPolicies map[string][]string `json:"contentSecurityPolicies"`
	PermissionPolicy        map[string][]string `json:"permissionPolicy"`
	IframeAllow             map[string][]string `json:"iframeAllow"`
	State                   *pluginState        `json:"state"`
	Slots                   []uiPluginSlot      `json:"slots"`
}

// uiPluginBuildConfig is local CLI-only config and never sent to backend.
type uiPluginBuildConfig struct {
	OutDir string `json:"outDir,omitempty"`
	Port   *int   `json:"port,omitempty"`
}
