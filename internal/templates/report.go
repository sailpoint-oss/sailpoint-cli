package templates

const builtInReportTemplates = `[
  {
    "name": "provisioning-and-security",
    "description": "All account unlocks in the tenant for a given time range",
    "variables": [{ "name": "days", "prompt": "Days before today" }],
    "queries": [
      {
        "queryString": "(type:provisioning AND created:[now-{{days}}d TO now])",
        "queryTitle": "Provisioning Events for the last {{days}} days"
      },
      {
        "queryString": "(USER_UNLOCK_PASSED AND created:[now-{{days}}d TO now])",
        "queryTitle": "User Unlocks for the last {{days}} days"
      }
    ]
  }
]`
