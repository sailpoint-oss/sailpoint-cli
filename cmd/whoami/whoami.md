==Long==
# Whoami

Show how the CLI is authenticated against SailPoint Identity Security: the active environment, the auth type (PAT or OAuth), where each setting came from (environment variables, config file or keychain), and what the current access token contains, such as the user, tenant, pod, scopes and expiry.

Any `SAIL_*` environment variables that are set are listed, because they override your saved config without telling you. `whoami` warns when one replaces a different saved value. With a PAT, it also looks up the token in your tenant and shows its name, owner, created date and expiry, with a warning if it's over a year old or about to expire.

`whoami` never opens a browser. It uses the cached token, gets a new PAT token or refreshes an OAuth token when that needs no prompt, and otherwise reports that you're not logged in.

====

==Example==
## Show the current login

```bash
sail whoami
```

## Show the login for a specific environment

```bash
sail whoami --env {environment-name}
```

## Print the result as JSON

```bash
sail whoami --json
```
====
