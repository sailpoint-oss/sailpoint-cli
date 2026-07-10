==Long==
# Init

Get a UI plugin project to a clean starting point. `init` has two flows on one command:

- **New workspace** (default): scaffolds a new Angular plugin workspace from the SailPoint UI plugin templates into a new directory named after the plugin alias.
- **Existing workspace** (`--path <dir>`): makes an existing project SDK-ready by generating a valid `sp-ui-plugin.json` and dropping the plugin guide, without touching unrelated files.

## Inputs

`init` prompts for the minimum inputs, each of which has an equivalent flag so the command can run headless in CI/CD. Providing `--name` (or the positional plugin name) skips the interactive prompts.

- **Plugin Name** — the display name (`--name`, or the positional argument).
- **Plugin Alias** — the tenant-unique key (`--alias`); defaults to a slug of the name. The alias is validated against your tenant immediately and the command fails fast if it is already taken or invalid.
- **Build Output Directory** — required when attaching with `--path` (`--out-dir`).
- **Port** — the local dev server port when attaching with `--path` (`--port`, default 3000).

The alias validation requires an authenticated CLI; `init` fails fast if no authenticated client is available.

## What init does not do

`init` does not install dependencies, build, or register the plugin, and it does not modify your `package.json` dependencies. The SDK ships with the template for new workspaces; for `--path` install it per the generated guide. After `init`, run `npm install` and `sail ui-plugins create` yourself.
====

==Example==

```bash
# Scaffold a new Angular workspace (interactive prompts for name/alias)
SAIL_EXPERIMENTAL_UI_PLUGINS=1 sail ui-plugins init

# Scaffold headlessly (no prompts)
SAIL_EXPERIMENTAL_UI_PLUGINS=1 sail ui-plugins init "My Plugin" --alias my-plugin

# Attach the SDK to an existing project
SAIL_EXPERIMENTAL_UI_PLUGINS=1 sail ui-plugins init "My Plugin" \
  --path ./existing-app --out-dir ./dist/app --port 3000

# Overwrite plugin files init manages if they already exist
SAIL_EXPERIMENTAL_UI_PLUGINS=1 sail ui-plugins init "My Plugin" --path ./existing-app --out-dir ./dist/app --force
```

====
