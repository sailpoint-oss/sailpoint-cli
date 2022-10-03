# Usage

Command `conn` is short for `connectors`. Both of the following commands work and they work the exact same way

```shell
$ sp conn list
$ sp connectors list
```

> Note that for all invoke commands, the version flag `-v` is optional. If not provided, the CLI will run against the version pointed by the `latest` tag.

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