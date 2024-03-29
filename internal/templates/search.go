package templates

const builtInSearchTemplates = `[
  {
    "name": "all-account-unlocks",
    "description": "All account unlocks in the tenant for a given time range",
    "variables": [{ "name": "days", "prompt": "Days before today" }],
    "searchQuery": {
      "indices": ["events"],
      "queryType": null,
      "queryVersion": null,
      "query": {
        "query": "(USER_UNLOCK_PASSED AND created:[now-{{days}}d TO now])"
      },
      "sort": [],
      "searchAfter": []
    }
  },
  {
    "name": "all-provisioning-events",
    "description": "All provisioning events in the tenant for a given time range",
    "variables": [{ "name": "days", "prompt": "Days before today" }],
    "searchQuery": {
      "indices": ["events"],
      "queryType": null,
      "queryVersion": null,
      "query": {
        "query": "(type:provisioning AND created:[now-{{days}}d TO now])"
      },
      "sort": [],
      "searchAfter": []
    }
  },
  {
    "name": "all-provisioning-events-90-days",
    "description": "All provisioning events in the tenant for a given time range",
    "variables": [],
    "searchQuery": {
      "indices": ["events"],
      "queryType": null,
      "queryVersion": null,
      "query": {
        "query": "(type:provisioning AND created:[now-90d TO now])"
      },
      "sort": [],
      "searchAfter": []
    }
  }
]`
