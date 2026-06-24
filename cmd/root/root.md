==Long==
# SailPoint CLI
The SailPoint CLI allows you to administer your Identity Security Cloud tenant from the command line.

## Common workflows

- `sail auth login` - sign in to the active environment
- `sail env list` - show configured tenants and authentication types
- `sail env create <name>` - add a tenant environment
- `sail identity list` - inspect identities
- `sail source list` - inspect sources
- `sail account list` - inspect accounts
- `sail role list` - inspect roles
- `sail access-profile list` - inspect access profiles
- `sail entitlement list` - inspect entitlements
- `sail access-request list` - review access request status
- `sail api get /v2024/...` - call an Identity Security Cloud API endpoint
- `sail workflow list` - inspect workflows
- `sail transform list` - inspect transforms
- `sail search query` - run saved or ad hoc searches
- `sail spconfig` - export, import, and monitor tenant configuration
- `sail connector` - build and operate connector projects

## Output and automation

Use `--output json`, `--output yaml`, or `--json` for machine-readable output where commands support structured results. User-facing logs and warnings are written to stderr so stdout can be piped to tools such as `jq`.

For CI, configure PAT credentials with environment variables instead of the local keyring: `SAIL_BASE_URL`, `SAIL_AUTHTYPE=pat`, `SAIL_CLIENT_ID`, and `SAIL_CLIENT_SECRET`.

Navigate to the [CLI Documentation](https://developer.sailpoint.com/docs/tools/cli) for full command documentation.

====

==Example==
```bash
sail 
sail --output json env list
sail env use production
printf '%s' "$SAIL_CLIENT_SECRET" | sail auth pat set --client-id "$SAIL_CLIENT_ID" --client-secret-stdin
sail identity list --filter 'name sw "a"'
sail role entitlements <role-id>
sail api get /v2024/identities --query limit=10 --pretty
```
====