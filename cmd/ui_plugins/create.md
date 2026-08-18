==Long==
# Create

Registers a UI plugin instance in the current tenant from the workspace manifest at `./sp-ui-plugin.json`.

The manifest is validated locally first (the same structural checks as `validate-manifest`); only the `manifest` section is sent to UMS — the local `build` section is never transmitted. The workspace alias becomes the instance identifier, and an alias already in use in the tenant is reported as a conflict.

## Visibility overrides

Both overrides apply per slot and replace any `restrictToUsers` declared in the manifest:

- `--restrict-to-users` adds the given user identity GUIDs to every slot.
- `--private` adds your own identity (resolved from the active token) to every slot.

When both are supplied, each slot's `restrictToUsers` is the de-duplicated union of your identity and the supplied GUIDs.

## Dry run

`--dry-run` validates the manifest, applies any overrides, and prints the exact payload that would be sent — without creating the instance. It also performs a read-only alias availability check against the tenant when possible; an alias that is already taken or invalid is reported, while an inconclusive check (e.g. connectivity or access) is noted without failing.

## Local dev document headers

On a successful create (not `--dry-run`), when the backend returns `devDocumentHeaders`, the CLI writes the `Content-Security-Policy` and `Permissions-Policy` values into `./angular.json` at `projects.<alias>.architect.serve.options.headers` (creating the `headers` object if needed). Values are copied as returned by the backend, including an empty `Permissions-Policy` when present. Other `serve.options` keys (such as HTTPS settings) are preserved. `sp-ui-plugin.json` is not modified for dev headers.

The Angular project patched is resolved from the manifest alias: the CLI looks for `projects.<alias>` first (matching the project name `init` sets during scaffolding). If that key is missing but `angular.json` defines exactly one project, that sole project is updated instead. If multiple projects are defined and none matches the alias, create still succeeds with a warning on stderr that `angular.json` could not be updated; local dev CSP may remain stale until you rename the project to match the alias or add the headers manually under the correct project.

If `./angular.json` is missing (for example, a non-Angular workspace attached with `init --path`), a note is printed and create still succeeds. When headers change, a restart reminder is printed to stderr — restart `ng serve` / `npm start` before testing in ISC.

Use `--json` to inspect the raw backend response, including `devDocumentHeaders`; header patching still runs as a side effect unless backend omits that field.

## Output

On success a short confirmation with the new plugin instance ID and alias is printed. Use `--json` to print the raw UMS response instead.
====

==Example==

```bash
# Create from ./sp-ui-plugin.json
SAIL_EXPERIMENTAL_UI_PLUGINS=1 sail ui-plugins create

# Preview the payload (and check alias availability) without creating
SAIL_EXPERIMENTAL_UI_PLUGINS=1 sail ui-plugins create --dry-run

# Restrict the plugin to yourself on every slot
SAIL_EXPERIMENTAL_UI_PLUGINS=1 sail ui-plugins create --private

# Restrict the plugin to specific users on every slot
SAIL_EXPERIMENTAL_UI_PLUGINS=1 sail ui-plugins create --restrict-to-users 2c9180...a1,2c9180...b2

# Print the raw UMS response on success
SAIL_EXPERIMENTAL_UI_PLUGINS=1 sail ui-plugins create --json
```

====
