==Long==
# Disable

Disables a UI plugin instance from the current tenant, identified by either its alias or its plugin ID.

Before disabling, the command looks up the instance and prints a summary (alias, name, plugin ID, created date, and slots) for you to confirm. If the instance has an active asset bundle (a live deployment), a warning is shown.

## Identifying the instance

The single argument accepts either an alias or a plugin ID; the type is detected automatically by its shape:

- A value in UUID form is treated as a plugin ID.
- Anything else is treated as an alias and resolved to its plugin ID first.

A UUID is technically also a valid alias. If you register an alias that looks like a UUID, this command will treat that argument as a plugin ID and report "not found" — it will never disable a different instance. Such an instance can still be disabled by passing its plugin ID.

## Confirmation and --force

Disabling requires confirmation (the prompt defaults to No). `--force` (`-F`) skips the confirmation prompt only — it does not bypass safety checks:

- If an alias resolves to more than one instance, the command lists the conflicting plugin IDs and stops, even with `--force`. Re-run with a specific plugin ID.
- A missing instance is always reported as an error.

## Output

On success a confirmation with the disabled plugin ID (and alias, when known) is printed. Use `--json` to print the disabled plugin instance as JSON instead.
====

==Example==

```bash
# Disable by alias (prompts for confirmation)
SAIL_EXPERIMENTAL_UI_PLUGINS=1 sail ui-plugins disable access-request-plugin

# Disable by plugin ID
SAIL_EXPERIMENTAL_UI_PLUGINS=1 sail ui-plugins disable 2c918085-7a1e-1b2c-817a-1e1b2c000000

# Skip the confirmation prompt
SAIL_EXPERIMENTAL_UI_PLUGINS=1 sail ui-plugins disable access-request-plugin --force

# Print the disabled instance as JSON (scripting)
SAIL_EXPERIMENTAL_UI_PLUGINS=1 sail ui-plugins disable access-request-plugin --force --json
```

====
