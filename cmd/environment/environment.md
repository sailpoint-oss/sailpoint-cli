==Long==
# Environment

Configure SailPoint IdentityNow environments for the CLI
====

==Example==
## Adding an new environment

You can add new environments by supplying a name that does not already exist. The CLI will prompt you for the Tenant URL and API URL.

```bash
sail environment {environment-name} 
```


## Switching environments

You can switch between environments by supplying the name of an existing environment you want to switch to.

```bash
sail environment {environment-name}
```  

## Removing an environment

You can remove an environment by supplying the name of an existing environment you want to remove in combination with the `--erase` flag.

```bash
sail environment {environment-name} --erase
```

## Overwriting an environment

You can overwrite an environment by supplying the name of an existing environment you want to overwrite in combination with the `--overwrite` flag.

```bash
sail environment {environment-name} --overwrite
```

## Showing an environment

You can print an environment by supplying the name of an existing environment you want to print in combination with the `--show` flag.

```bash
sail environment {environment-name} --show
```
====