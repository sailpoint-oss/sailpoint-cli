package scopes

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

//go:generate go run -tags scopesgen ./gen -spec https://github.com/sailpoint-oss/api-specs/releases/latest/download/deref-sailpoint-api.json -out data/sailpoint-scopes.json

// SchemaVersion is the scope manifest format version this package understands.
const SchemaVersion = 1

// Language identifies the SDK that a call was made through.
type Language string

const (
	LangGo         Language = "go"
	LangPython     Language = "python"
	LangTypeScript Language = "typescript"
	LangAngular    Language = "angular"
	LangPowerShell Language = "powershell"
)

// Languages lists every supported SDK language in display order.
var Languages = []Language{LangGo, LangPython, LangTypeScript, LangAngular, LangPowerShell}

// Manifest is the sailpoint-scopes.json file shipped with each SDK, and the
// embedded snapshot used when no installed SDK manifest is found.
type Manifest struct {
	SchemaVersion int                  `json:"schemaVersion"`
	SDK           string               `json:"sdk,omitempty"`
	SDKVersion    string               `json:"sdkVersion,omitempty"`
	Source        string               `json:"source,omitempty"`
	Operations    map[string]Operation `json:"operations"`
}

// Operation holds the security requirements of one API operation.
//
// UserAuth and ApplicationAuth are "any one of" scope lists. A nil list means
// that the auth type cannot call the operation. An empty list means that the
// auth type can call it without any scope.
type Operation struct {
	Resource        string              `json:"resource,omitempty"`
	Tag             string              `json:"tag,omitempty"`
	SDKName         string              `json:"sdkName,omitempty"`
	SDKNames        map[Language]string `json:"sdkNames,omitempty"`
	HTTPMethod      string              `json:"httpMethod"`
	Path            string              `json:"path"`
	UserAuth        []string            `json:"userAuth"`
	ApplicationAuth []string            `json:"applicationAuth"`
	DefaultScope    bool                `json:"defaultScope,omitempty"`
	Public          bool                `json:"public,omitempty"`
	UserLevels      []string            `json:"userLevels,omitempty"`
}

// ScopesFor returns the scope alternatives of op for the given auth type.
func (op Operation) ScopesFor(auth AuthType) []string {
	if auth == AuthApplication {
		return op.ApplicationAuth
	}
	return op.UserAuth
}

// Load reads and parses a scope manifest from disk.
func Load(path string) (*Manifest, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	m, err := Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return m, nil
}

// Parse decodes a scope manifest and checks its schema version.
func Parse(raw []byte) (*Manifest, error) {
	var m Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("invalid scope manifest: %w", err)
	}
	if m.SchemaVersion != SchemaVersion {
		return nil, fmt.Errorf("unsupported scope manifest schemaVersion %d (supported: %d)", m.SchemaVersion, SchemaVersion)
	}
	return &m, nil
}

// NameIndex maps the SDK method or cmdlet name for lang to its operationId.
// PowerShell names are stored in lower case because PowerShell is case
// insensitive.
func (m *Manifest) NameIndex(lang Language) map[string]string {
	index := make(map[string]string, len(m.Operations))
	for id, op := range m.Operations {
		name := op.SDKName
		if name == "" {
			name = op.SDKNames[snapshotKey(lang)]
		}
		if name == "" {
			continue
		}
		if lang == LangPowerShell {
			name = strings.ToLower(name)
		}
		index[name] = id
	}
	return index
}

// snapshotKey returns the sdkNames key for lang. The Angular SDK uses the
// same method names as the TypeScript SDK.
func snapshotKey(lang Language) Language {
	if lang == LangAngular {
		return LangTypeScript
	}
	return lang
}
