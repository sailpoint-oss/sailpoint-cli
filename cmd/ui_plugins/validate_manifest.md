==Long==
# Validate Manifest

Performs an offline structural validation of `./sp-ui-plugin.json` in the current working directory.

## What is checked

- JSON shape, required fields, and known field types
- Rejection of unknown fields (strict schema)
- CLI config schema version support
- Presence of the required policy objects (`contentSecurityPolicies`, `permissionPolicy`, `iframeAllow`); an empty object `{}` is accepted
- Token format of each entry in the optional `iframeSandbox` array: lowercase ASCII letters, digits, and hyphens, with a hyphen only between characters (for example `allow-popups`); an empty array is accepted

## What is not checked

This command does not contact UMS or validate tenant-specific rules, including:

- Alias availability or format rules enforced by the backend
- Slot registry membership and occupancy limits
- Security policy directive names/values and allowlists (CSP, permission policy, iframe allow) — only that the policy objects are present
- The set of accepted `iframeSandbox` tokens — UMS enforces which tokens are permitted; the CLI checks only each token's format
- Capability lists, user GUID restrictions, and related business rules

Use `sail ui-plugins create` or `update` for full backend validation.

====

==Example==

```bash
sail ui-plugins validate-manifest
sail ui-plugins validate
```

====
