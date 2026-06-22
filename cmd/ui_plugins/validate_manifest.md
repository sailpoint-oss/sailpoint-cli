==Long==
# Validate Manifest

Performs an offline structural validation of `./sp-ui-plugin.json` in the current working directory.

## What is checked

- JSON shape, required fields, and known field types
- Rejection of unknown fields (strict schema)
- CLI config schema version support

## What is not checked

This command does not contact UMS or validate tenant-specific rules, including:

- Alias availability or format rules enforced by the backend
- Slot registry membership and occupancy limits
- Security policy baselines (CSP, permission policy, iframe allow)
- Capability lists, user GUID restrictions, and related business rules

Use `sail ui-plugins create` or `update` for full backend validation.

====

==Example==

```bash
SAIL_EXPERIMENTAL_UI_PLUGINS=1 sail ui-plugins validate-manifest
SAIL_EXPERIMENTAL_UI_PLUGINS=1 sail ui-plugins validate
```

====
