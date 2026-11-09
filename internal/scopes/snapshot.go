//go:build !scopesgen

package scopes

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
)

const snapshotFile = "data/sailpoint-scopes.json"

// The whole data directory is embedded so that the build does not fail when
// the snapshot is missing. The directory holds a committed README.md, and the
// snapshot itself is generated and not committed. Snapshot reports a missing
// file with instructions instead of a compiler error.
//
//go:embed data
var snapshotFS embed.FS

// ErrNoSnapshot means that the CLI was built without the scope snapshot.
var ErrNoSnapshot = errors.New("the scope snapshot was not embedded in this build: " +
	"run `go generate ./internal/scopes/...` (or `make generate`) to download it, then build again")

// Snapshot returns the scope manifest embedded in the CLI at build time.
func Snapshot() (*Manifest, error) {
	raw, err := snapshotFS.ReadFile(snapshotFile)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, ErrNoSnapshot
	}
	if err != nil {
		return nil, fmt.Errorf("embedded scope snapshot: %w", err)
	}
	m, err := Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("embedded scope snapshot: %w", err)
	}
	return m, nil
}
