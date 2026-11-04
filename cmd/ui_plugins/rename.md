==Long==
# Rename

Renames a UI plugin's display name and/or alias, keeping your local workspace and the plugin instance in your tenant in sync. Without it, a rename means editing `sp-ui-plugin.json` by hand and remembering to push the change.

Run from a plugin workspace — the directory that contains `./sp-ui-plugin.json`.

## What it changes

- `--name` updates the English display name (`manifest.name.en`).
- `--alias` updates the tenant-unique alias (`manifest.alias`), stored lowercase.

Only those two fields are touched. Your description, other locales, the local `build` section, and every other project file are left unchanged. Static assets are not uploaded — use `upload` for that.

## Prompts

Omit a flag and you are prompted for it, defaulting to the current value — press Enter to keep it. A value passed as a flag is not prompted for. If neither the name nor the alias actually changes, nothing is written and the tenant is not called.

## How a changed alias is applied

The current alias is resolved against your tenant first:

- If the plugin has not been created yet, the file is updated and you are pointed at `create`.
- If it already exists, a changed alias is checked for availability, the plugin instance is updated, and only then is the file written — so a failed update leaves your workspace untouched.

Changing the alias changes your local development URL, so re-run `link` afterwards if you had already linked.

Requires the `idn:plugins-ui:update` right.
====

==Example==
```bash
# Prompt for the new name and alias, each defaulting to the current value
sail ui-plugins rename

# Rename non-interactively
sail ui-plugins rename --name "Access Requests" --alias access-requests

# Change only the alias
sail ui-plugins rename --alias access-requests
```
====
