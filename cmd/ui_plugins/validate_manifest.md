==Long==
# Validate Manifest

Performs an offline structural validation of `./sp-ui-plugin.json` in the current working directory.

## What is checked

- JSON shape, required fields, and known field types
- Rejection of unknown fields (strict schema)
- CLI config schema version support
- Presence of the required policy objects (`contentSecurityPolicies`, `permissionPolicy`, `iframeAllow`); an empty object `{}` is accepted
- The optional `iframeSandbox` field is a JSON array decoded with the same string-slice rules as policy values

## What is not checked

This command does not contact UMS or validate tenant-specific rules, including:

- Alias availability or format rules enforced by the backend
- Slot registry membership and occupancy limits
- Security policy directive names/values and allowlists (CSP, permission policy, iframe allow, iframe sandbox) — only their JSON structure is checked
- Capability lists, user GUID restrictions, and related business rules

UMS validates sandbox token values and policy expressions, including the single-quoted `"'src'"` value used by `iframeAllow` and `"'self'"` used by `permissionPolicy`. The CLI forwards these strings as written and does not rewrite them.

For example, an optional empty sandbox list is structurally valid:

```json
{
  "permissionPolicy": {"camera": ["'self'"]},
  "iframeAllow": {"camera": ["'src'"]},
  "iframeSandbox": []
}
```

Use `sail ui-plugins create` or `push-manifest` for full backend validation.

====

==Example==

```bash
sail ui-plugins validate-manifest
sail ui-plugins validate
```

====
