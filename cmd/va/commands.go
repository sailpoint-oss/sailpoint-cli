package va

import (
	_ "embed"
)

// TroubleshootingScript is the SailPoint STUNT troubleshooting script, embedded at build time so the
// CLI never executes remotely fetched content on a VA.
//
// Source: https://github.com/sailpoint-oss/colab-stunt-script/blob/9b560a3af7697ff70d6b49c18210ef478c0febf1/stunt.sh (v2.4.2)
//
// To update, copy the reviewed stunt.sh from a pinned commit of sailpoint-oss/colab-stunt-script into
// scripts/stunt.sh and update TroubleshootingScriptSHA256 and the source commit above.
//
//go:embed scripts/stunt.sh
var TroubleshootingScript []byte

// TroubleshootingScriptSHA256 is the SHA-256 digest of the reviewed scripts/stunt.sh.
const TroubleshootingScriptSHA256 = "755fe5dfcac77fb4b5429e0c877ca5631f142a5d1763c6a7bcc5c2ff6b24b7ae"

// TroubleshootingArchiveGlob matches the zip archive STUNT writes when run with default options.
const TroubleshootingArchiveGlob = "/home/sailpoint/logs.*.zip"

const UpdateCommand = "sudo update_engine_client -check_for_update"
const RebootCommand = "sudo reboot"
