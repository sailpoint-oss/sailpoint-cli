==Long==
# Unlink

Removes your local development link for a UI plugin so ISC stops loading your local code and returns to serving the deployed plugin. Run this from within an initialized plugin workspace; the alias is read from `./sp-ui-plugin.json`, and the active tenant is resolved from your CLI authentication.

Unlinking affects only your own local dev override on the plugin instance — it never changes the deployed plugin or other developers' overrides. It is the inverse of `link`.

## Idempotent

Unlinking is safe to run whether or not a link currently exists. If you have no active local dev override for this plugin, the command reports success without making a change.

## Output

On success a short confirmation is printed. Backend failures are surfaced with actionable detail.
====

==Example==

```bash
# Remove your local dev link for the plugin in this workspace
SAIL_EXPERIMENTAL_UI_PLUGINS=1 sail ui-plugins unlink
```

====
