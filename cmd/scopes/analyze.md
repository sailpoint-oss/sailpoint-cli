==Long==
# Analyze

Scans the current directory and all directories below it for calls to the SailPoint SDKs. Then it lists the smallest set of OAuth scopes that those calls need.

Files and directories in `.gitignore` are not scanned. Dependency directories such as `node_modules`, `vendor` and `.venv` are never scanned.

## Supported SDKs

The Go, Python, TypeScript, Angular and PowerShell SDKs, with version-suffixed methods such as `listAccountsV1`. Direct HTTP calls are not detected.

## Scope data

The command reads `sailpoint-scopes.json` from the SDK that your project installed, so the scopes match your SDK version. If no installed SDK has this file, the command uses the scope data embedded in the CLI and prints a warning.

## Choosing scopes

Many operations accept any one of several scopes, for example `idn:accounts:read` or `idn:accounts:manage`. The command chooses one scope that covers as many operations as possible, and prefers read scopes. The `Scopes` column and `--json` show all alternatives.
====

==Example==

```bash
# Analyze the current directory
sail scopes analyze

# Analyze another directory with application (client credentials) scopes
sail scopes analyze --path ./my-app --auth-type application

# Print the scopes as a comma-separated list
sail scopes analyze --scopes-only

# Write the scopes into apiScopes in ./sp-ui-plugin.json
sail scopes analyze --update-manifest
```

====
