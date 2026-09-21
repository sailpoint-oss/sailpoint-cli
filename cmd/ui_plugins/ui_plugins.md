==Long==
# UI Plugins
Manage UI Plugin workflows in Identity Security Cloud.

## Typical flow

**Local development**
1. `init` — scaffold a new workspace (or prepare an existing project with `--path`).
2. Install your project's dependencies.
3. `create` — register the plugin instance in your tenant from `./sp-ui-plugin.json`.
4. `link` — bind your local dev server to your identity and print the developer URL.
5. Start your local dev server, then open the developer URL in ISC to load your local code.

**Live deployment**
1. Build your project with your framework's native tooling (e.g. `ng build`).
2. `upload` — deploy the compiled static assets.
3. `push-manifest` — deploy manifest/configuration changes only.

Run any command with `--help` for details. `validate-manifest` (alias: `validate`)
performs offline structural validation of `./sp-ui-plugin.json` without calling UMS.
====

==Example==
```bash
sail ui-plugins
sail ui-plugins init
```
====
