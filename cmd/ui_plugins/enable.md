==Long==
# Enable

Enables a UI plugin instance from the current tenant, identified by either its alias or its plugin ID.

## Identifying the instance

The single argument accepts either an alias or a plugin ID; the type is detected automatically by its shape:

- A value in UUID form is treated as a plugin ID.
- Anything else is treated as an alias and resolved to its plugin ID first.
- If an alias resolves to more than one instance, the command lists the conflicting plugin IDs and stops. Re-run with a specific plugin ID.
- A missing instance is always reported as an error.

A UUID is technically also a valid alias. If you register an alias that looks like a UUID, this command will treat that argument as a plugin ID and report "not found" — it will never enable a different instance. Such an instance can still be enabled by passing its plugin ID.

## Output

On success a confirmation with the enabled plugin ID (and alias, when known) is printed. Use `--json` to print the enabled plugin instance as JSON instead.
====

==Example==

```bash
# Enable by alias
SAIL_EXPERIMENTAL_UI_PLUGINS=1 sail ui-plugins enable access-request-plugin

# Enable by plugin ID
SAIL_EXPERIMENTAL_UI_PLUGINS=1 sail ui-plugins enable 2c918085-7a1e-1b2c-817a-1e1b2c000000

# Print the enabled instance as JSON (scripting)
SAIL_EXPERIMENTAL_UI_PLUGINS=1 sail ui-plugins enable access-request-plugin --json
```

====
