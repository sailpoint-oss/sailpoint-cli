==Long==
# List

Lists the UI plugin instances registered in the current tenant.

All instances are returned — the command pages through the backend internally, so there is no need to specify a limit or offset. Results are shown as a table sorted by alias, with the columns Alias, Id, Name, and Created.

## Output

By default a table is printed; an empty tenant prints `No plugin instances found.` Long author-entered values (alias, name) are truncated with an ellipsis to keep the table readable — use `--json` to print the raw, untruncated plugin instance list as a JSON array (useful for scripting).
====

==Example==

```bash
# List all plugin instances in the current tenant
SAIL_EXPERIMENTAL_UI_PLUGINS=1 sail ui-plugins list

# Print the raw list as JSON
SAIL_EXPERIMENTAL_UI_PLUGINS=1 sail ui-plugins list --json
```

====
