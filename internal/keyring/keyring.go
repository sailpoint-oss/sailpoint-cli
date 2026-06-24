package keyring

import (
	"sync"
	"time"

	gokeyring "github.com/zalando/go-keyring"
)

// Get retrieves a value from the system keyring for the given service and environment.
func Get(service, env string) (string, error) {
	return gokeyring.Get(service, env)
}

// Set stores a value in the system keyring for the given service and environment.
func Set(service, env, value string) error {
	return gokeyring.Set(service, env, value)
}

// Delete removes a value from the system keyring for the given service and environment.
func Delete(service, env string) error {
	return gokeyring.Delete(service, env)
}

// GetTime retrieves a time.Time value from the keyring (stored as RFC3339 string).
func GetTime(service, env string) (time.Time, error) {
	s, err := Get(service, env)
	if err != nil {
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339, s)
}

// SetTime stores a time.Time value in the keyring as an RFC3339 string.
func SetTime(service, env string, t time.Time) error {
	return Set(service, env, t.Format(time.RFC3339))
}

var (
	secretsSupported     *bool
	secretsSupportedOnce sync.Once
)

// IsSupported checks whether the system keyring is functional.
// The result is cached after the first call.
func IsSupported() bool {
	secretsSupportedOnce.Do(func() {
		supported := testKeyring()
		secretsSupported = &supported
	})
	return *secretsSupported
}

func testKeyring() bool {
	const svc = "sailpoint-cli.test"
	const user = "test"
	const secret = "test-secret"

	if err := gokeyring.Set(svc, user, secret); err != nil {
		return false
	}
	val, err := gokeyring.Get(svc, user)
	if err != nil || val != secret {
		return false
	}
	_ = gokeyring.Delete(svc, user)
	return true
}
