package ui_plugins

// ManifestFileName is the workspace manifest file name, for callers outside
// this package.
const ManifestFileName = manifestFileName

// UpdateManifestAPIScopes replaces manifest.apiScopes in the workspace
// manifest at path. The file is validated before and after the change, and
// written in the same format as `ui-plugins init`.
func UpdateManifestAPIScopes(path string, apiScopes []string) error {
	cfg, err := loadAndValidateWorkspaceManifest(path)
	if err != nil {
		return err
	}
	cfg.Manifest.APIScopes = apiScopes
	return writeWorkspaceManifest(path, cfg)
}
