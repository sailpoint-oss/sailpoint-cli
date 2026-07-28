==Long==
# Push Manifest

Pushes your local manifest configuration to an existing UI plugin instance in the current tenant, from the workspace manifest at `./sp-ui-plugin.json`. This command is also available under the alias `update`.

Only the plugin configuration is updated — static assets are not uploaded (use `upload` for that). The manifest is validated locally first (the same structural checks as `validate-manifest`); only the `manifest` section is sent to UMS — the local `build` section is never transmitted. The workspace alias is resolved to its plugin instance in the tenant, and the entire manifest payload is sent, so any changed fields are applied.

The plugin must already exist (created with `create`). Payload validation is performed by UMS; validation errors are surfaced directly so you can correct the manifest.

## Visibility overrides

Both overrides apply per slot and replace any `restrictToUsers` declared in the manifest:

- `--restrict-to-users` adds the given user identity GUIDs to every slot.
- `--private` adds your own identity (resolved from the active token) to every slot.

When both are supplied, each slot's `restrictToUsers` is the de-duplicated union of your identity and the supplied GUIDs.

## Dry run

`--dry-run` validates the manifest, applies any overrides, and prints the exact payload that would be sent — without updating the instance. It also performs a read-only check that the workspace alias resolves to an existing instance; an alias that resolves to nothing (run `create` first) or to more than one instance is reported, while an inconclusive check (e.g. connectivity or access) is noted without failing.

## Output

On success a short confirmation with the plugin instance ID and alias is printed. Use `--json` to print the raw UMS response instead.
====

==Example==

```bash
# Push your manifest from ./sp-ui-plugin.json
SAIL_EXPERIMENTAL_UI_PLUGINS=1 sail ui-plugins push-manifest

# Also available under the update alias
SAIL_EXPERIMENTAL_UI_PLUGINS=1 sail ui-plugins update

# Preview the payload (and check the instance exists) without pushing
SAIL_EXPERIMENTAL_UI_PLUGINS=1 sail ui-plugins push-manifest --dry-run

# Restrict the plugin to yourself on every slot
SAIL_EXPERIMENTAL_UI_PLUGINS=1 sail ui-plugins push-manifest --private

# Restrict the plugin to specific users on every slot
SAIL_EXPERIMENTAL_UI_PLUGINS=1 sail ui-plugins push-manifest --restrict-to-users 2c9180...a1,2c9180...b2

# Print the raw UMS response on success
SAIL_EXPERIMENTAL_UI_PLUGINS=1 sail ui-plugins push-manifest --json
```

====
