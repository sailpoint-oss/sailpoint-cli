==Long==
# Export
Start an export job in Identity Security Cloud.

You can include or exclude the following valid types:
 - ACCESS_PROFILE
 - ACCESS_REQUEST_CONFIG
 - ATTR_SYNC_SOURCE_CONFIG
 - AUTH_ORG
 - CAMPAIGN_FILTER
 - FORM_DEFINITION
 - GOVERNANCE_GROUP
 - IDENTITY_OBJECT_CONFIG
 - IDENTITY_PROFILE
 - LIFECYCLE_STATE
 - NOTIFICATION_TEMPLATE
 - PASSWORD_POLICY
 - PASSWORD_SYNC_GROUP
 - PUBLIC_IDENTITIES_CONFIG
 - ROLE
 - RULE
 - SERVICE_DESK_INTEGRATION
 - SOD_POLICY
 - SOURCE
 - TRANSFORM
 - TRIGGER_SUBSCRIPTION
 - WORKFLOW
====

==Example==
```bash
sail spconfig export --include WORKFLOW --include SOURCE
sail spconfig export --include SOURCE --wait
sail spconfig export --include TRANSFORM --objectOptions '{
    "TRANSFORM": {
      "includedIds": [],
      "includedNames": [
        "Remove Diacritical Marks",
        "Common - Location Lookup"
      ]
    }
  }'
```
====