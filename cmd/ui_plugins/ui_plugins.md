==Long==
# UI Plugins
Manage UI Plugin workflows in Identity Security Cloud.

This command group is currently experimental and hidden from default command discovery.
Enable it for development with:

```bash
export SAIL_EXPERIMENTAL_UI_PLUGINS=1
```

Planned workflow commands:
- init
- create
- link
- push-manifest (alias: update)
- upload
- list
- delete
- validate-manifest (alias: validate) — offline structural validation of `./sp-ui-plugin.json`; does not call UMS
====

==Example==
```bash
sail ui-plugins
SAIL_EXPERIMENTAL_UI_PLUGINS=1 sail ui-plugins init
```
====
