```
   _____       _ _ _____      _       _      _____ _      _____ 
  / ____|     (_) |  __ \    (_)     | |    / ____| |    |_   _|
 | (___   __ _ _| | |__) |__  _ _ __ | |_  | |    | |      | |  
  \___ \ / _` | | |  ___/ _ \| | '_ \| __| | |    | |      | |  
  ____) | (_| | | | |  | (_) | | | | | |_  | |____| |____ _| |_ 
 |_____/ \__,_|_|_|_|   \___/|_|_| |_|\__|  \_____|______|_____|
                                                                
```
The SailPoint Command Line Interface (CLI) makes it easy to interact with SailPoint's SaaS Platform in a programmatic way.  Many functions that use to be accomplished through tools like Postman or from custom scripts can now be done directly on the command line with minimal setup.

## Install

Installation of the CLI requires [Golang](https://go.dev/doc/install) version 1.17 or above.

### MacOS and Linux

Open your terminal app, navigate to the project directory, and run the following command.

```shell
make install
```

After that, make sure you can run the `sail` command.

```shell
sail -h
```

### Windows

Open PowerShell, navigate to the project directory, and run the following command.

```shell
go build -o "C:\Program Files\sailpoint\sail.exe"
```

After that, add the following directory to the system PATH parameter. You can find instructions on how to do this from [this article](https://medium.com/@kevinmarkvi/how-to-add-executables-to-your-path-in-windows-5ffa4ce61a53). This will only need to be done the first time you install the CLI.

```text
C:\Program Files\sailpoint
```

Once installed, make sure PowerShell can run the `sail` command.

```shell
sail -h
```

## Configuration

Create a [personal access token](https://developer.sailpoint.com/docs/authentication.html#personal-access-tokens), which will be used to authenticate the SP CLI to your IdentityNow tenant.

Create a configuration file in your home directory to save your credentials.

On Linux/Mac, run:

```shell
mkdir ~/.sp
touch ~/.sp/config.yaml
```

On Windows PowerShell, run:

```powershell

```

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
