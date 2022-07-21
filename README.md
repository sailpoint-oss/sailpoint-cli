# SP CLI (BETA)

SailPoint CLI

## Install
Installation of cli requires Golang. Make sure Golang is installed on your system with version 1.17 or above.

### MacOS and Linux
Run the following make command.
```shell
$ make install
```

After that, make sure you can run the `sp` command.
```shell
$ sp -h
```

### Windows
Install cli using the following command.
```shell
$ go build -o "C:\Program Files\sp-cli\sp"
```

After that, add the following directory to the system PATH parameter. This will only need to be done the first time you install the cli.
```
C:\Program Files\sp-cli
```

Once installed, make sure to use a bash-like shell to run cli commands. You can use MinGW or Git Bash. Make sure you can run the `sp` command.
```shell
$ sp -h
```


## Configuration

Create personal access token @ https://{org}.identitysoon.com/ui/d/user-preferences/personal-access-tokens

Create a config file at "~/.sp/config.yaml"

```yaml
baseURL: https://{org}.api.cloud.sailpoint.com # or baseURL: https://localhost:7100
tokenURL: https://{org}.api.cloud.sailpoint.com/oauth/token
clientSecret: [clientSecret]
clientID: [clientID]
```

You may also specify the config as environment variables:

```shell
$ SP_CLI_BASEURL=http://localhost:7100 \
  SP_CLI_TOKENURL=http://{org}.api.cloud.sailpoint.com \
  SP_CLI_CLIENTSECRET=xxxx sp conn list
```

This can useful for cases like CI pipelines to avoid having to write the config
file.

## Usage

Note that for all invoke commands, the version flag `-v` is optional. If not provided, the cli will run against the version pointed by the `latest` tag.

```shell
$ sp conn help
$ sp conn init [connectorProjectName]
$ sp conn create [connectorAlias]
$ sp conn update -c [connectorID] -a [connectorAlias]
$ sp conn list
$ sp conn upload -c [connectorID | connectorAlias] -f connector.zip
$ sp conn invoke test-connection -c [connectorID | connectorAlias] -p [config.json] -v [version]
$ sp conn invoke account-list -c [connectorID | connectorAlias] -p [config.json] -v [version]
$ sp conn invoke account-read [identity] -c [connectorID | connectorAlias] -p [config.json] -v [version]
$ sp conn invoke entitlement-list -t [entitlementType] -c [connectorID | connectorAlias] -p [config.json] -v [version]
$ sp conn invoke entitlement-read [identity] -t [entitlementType] -c [connectorID | connectorAlias] -p [config.json] -v [version]
$ sp conn tags create -c [connectorID | connectorAlias] -n [tagName] -v [version]
$ sp conn tags update -c [connectorID | connectorAlias] -n [tagName] -v [version]
$ sp conn tags list -c [connectorID | connectorAlias]
$ sp conn logs
$ sp conn logs tail
$ sp conn stats
```

### Command `conn` is short for `connectors`. Both of the following commands work and they work the exact same way

```shell
$ sp conn list
$ sp connectors list
```
