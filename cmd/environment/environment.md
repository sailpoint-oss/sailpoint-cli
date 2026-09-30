==Long==
# Environment

Configure SailPoint Identity Security Cloud environments for the CLI
====

==Example==
## Create a new environment

You can add new environments by calling `create` and supplying a name that does not already exist. The CLI will walk you through adding a new tenant configuration.

```bash
sail environment create {environment-name}
```

If no environment is provided the CLI will use your tenant name as the name of the environment.

```bash
sail environment create
```

When the CLI asks for the tenant name, you can also paste your tenant URL. This works for tenants that are not on the default `identitynow.com` domain. The CLI derives the environment name, the tenant URL, and the API URL from it. It then checks the API URL with the tenant. If the check passes, the CLI skips the URL confirmation prompts.

```bash
sail environment create
Tenant name or URL (ie: acme, or https://acme.identitynow-demo.com): () https://devrel-ga-25031.identitynow-demo.com/

✔ Found tenant at https://devrel-ga-25031.identitynow-demo.com (API: https://devrel-ga-25031.api.identitynow-demo.com)
```

## Switching environments

You can switch between environments by calling `use` and supplying the name of an existing environment you want to switch to.

```bash
sail environment use {environment-name}
```  

## Delete an environment

You can delete an environment by calling `delete` and supplying the name of an existing environment.

```bash
sail environment delete {environment-name}
```

If no environment is provided this command will delete the active environment.

```bash
sail environment delete
```

## Update an environment

You can update an environment by calling `update` and supplying the name of an existing environment.

```bash
sail environment update {environment-name}
```

If no environment is provided this command will delete the active environment.

```bash
sail environment update
```

## View an environment

You can print an environment by calling `show` supplying the name of an existing environment.

```bash
sail environment show {environment-name}
```

If no environment is provided this command will show the active environment.

```bash
sail environment show
```

====
