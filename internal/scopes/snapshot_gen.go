//go:build scopesgen

package scopes

import "errors"

// Snapshot is not available while the generator builds, because the
// generator writes the snapshot file that the normal build embeds.
func Snapshot() (*Manifest, error) {
	return nil, errors.New("scope snapshot is not embedded in scopesgen builds")
}
