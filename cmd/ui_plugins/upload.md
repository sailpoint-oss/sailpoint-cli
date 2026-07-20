==Long==
# Upload

Uploads the compiled UI plugin assets from a workspace to its plugin instance in the current tenant. Run this from within an initialized plugin workspace; the alias and configuration are read from `./sp-ui-plugin.json`, and the active tenant is resolved from your CLI authentication.

Uploading is a deploy step, not a build step. The command uploads the assets already present in the build output directory — it never runs your framework's build (e.g. `ng build`) for you. Compile the project with your native tooling first, then upload the result.

## Build output directory

The directory to upload is resolved in this order:

- `--out-dir` when provided.
- otherwise the `build.outDir` value from `sp-ui-plugin.json`.

If the resolved directory is missing or contains no assets, the command fails fast with a clear error rather than attempting to build the project.

## Deploy target (alias portability)

The workspace `alias` is the deterministic lookup key for the plugin instance, so the same command deploys to whichever tenant your CLI is currently authenticated against — run it in a staging context to update the staging instance, switch context to production and run the identical command to update production. This enables "write once, deploy anywhere" without tracking environment-specific plugin GUIDs locally.

On success the uploaded assets are hosted immutably behind the CDN and become the plugin instance's active asset bundle.

## Output

On success a short confirmation is printed. Backend and packaging failures are surfaced with actionable detail so authors can correct the workspace or retry.
====

==Example==

```bash
# Upload the compiled assets from sp-ui-plugin.json's build.outDir
SAIL_EXPERIMENTAL_UI_PLUGINS=1 sail ui-plugins upload

# Upload from an explicit directory, overriding build.outDir
SAIL_EXPERIMENTAL_UI_PLUGINS=1 sail ui-plugins upload --out-dir ./dist/my-app
```

====
