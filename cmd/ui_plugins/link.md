==Long==
# Link

Registers your local development server with a UI plugin instance so ISC loads your local code in the live tenant. Run this from within an initialized plugin workspace; the alias is read from `./sp-ui-plugin.json`, and the active tenant is resolved from your CLI authentication.

Linking binds your local dev server port to your own identity on the plugin instance (a per-developer override). It does not affect other developers or the plugin's deployed assets — it only changes what *you* see when you open the developer URL in ISC.

## Port resolution

The port registered as your local dev server is resolved in this order:

- `--port` when provided.
- otherwise the `build.port` value from `sp-ui-plugin.json`.
- otherwise the default dev server port (4200).

## Developer URL

On success the command prints a single fully qualified developer URL to stdout, for example `https://<tenant>/ui/plugin/<plugin-instance-id>?spPluginDev=<alias>`. Open it in a browser: `sp-renderer` verifies the override with UMS and, if you are authorized, loads your local code in the live tenant with a local-dev badge. The plugin URL is stable across links — re-running `link` with a different port updates the bound port without changing the URL.

The URL is the only thing written to stdout (informational messages go to stderr) so it can be piped or picked up by your terminal's link detection without truncation or masking.

## Output

The developer URL is printed to stdout; a short confirmation of the linked port is printed to stderr. Backend failures are surfaced with actionable detail.
====

==Example==

```bash
# Link using the port from sp-ui-plugin.json's build.port (or the 4200 default)
SAIL_EXPERIMENTAL_UI_PLUGINS=1 sail ui-plugins link

# Link an explicit port, overriding build.port
SAIL_EXPERIMENTAL_UI_PLUGINS=1 sail ui-plugins link --port 4300
```

====
