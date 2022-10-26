```
   _____       _ _ _____      _       _      _____ _      _____ 
  / ____|     (_) |  __ \    (_)     | |    / ____| |    |_   _|
 | (___   __ _ _| | |__) |__  _ _ __ | |_  | |    | |      | |  
  \___ \ / _` | | |  ___/ _ \| | '_ \| __| | |    | |      | |  
  ____) | (_| | | | |  | (_) | | | | | |_  | |____| |____ _| |_ 
 |_____/ \__,_|_|_|_|   \___/|_|_| |_|\__|  \_____|______|_____|
                                                                
```

The SailPoint Command Line Interface (CLI) makes it easy to interact with SailPoint's SaaS Platform in a programmatic way.  Many functions that use to be accomplished through tools like Postman or from custom scripts can now be done directly on the command line with minimal setup.

> **CAUTION:** The SailPoint CLI is currently in pre-production and undergoing heavy development.  Until the CLI reaches version 1.0.0, breaking changes may be introduced at any time while we work on refining the CLI.

## Contents

- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [Contributing](#contributing)
- [Code of Conduct](#code-of-conduct)

# Installation

## Prerequisites

Installation of the CLI requires [Golang](https://go.dev/doc/install) version 1.17 or above.

## MacOS and Linux

Open your terminal app, navigate to the project directory, and run the following command.

```shell
make install
```

After that, make sure you can run the `sail` command.

```shell
sail
```

## Windows

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
sail
```

# Configuration

Create a [personal access token](https://developer.sailpoint.com/idn/api/authentication#personal-access-tokens), which will be used to authenticate the SailPoint CLI to your IdentityNow tenant.

## Assisted configuration

Run the configure command to configure the CLI for your tenant.  This command will create a configuration file in your home directory to store your tenant's connection details.

```shell
sail configure
```

## Manual configuration

Alternatively, you can manually create a configuration file in your home directory.

On **Linux/Mac**, run:

```shell
mkdir ~/.sailpoint
touch ~/.sailpoint/config.yaml
```

On **Windows PowerShell**, run:

```powershell
New-Item -ItemType Directory -Path 'C:\Users\<username>\.sailpoint'
New-Item -ItemType File -Path 'C:\Users\<username>\.sailpoint\config.yaml' 
```

The `config.yaml` file should contain the following information.

```yaml
baseURL: https://{org}.api.identitynow.com # or baseURL: https://localhost:7100
tokenURL: https://{org}.api.identitynow.com/oauth/token
clientSecret: {clientSecret}
clientID: {clientID}
debug: false # Set to true for additional output
```

## Environment variable configuration

You may also specify environment variables for your configuration.  This can useful when using the CLI in an automated environment, like a CI/CD pipeline, where consuming the configuration from environment variables would be easier than creating the config file.  Environment variables take precedent over values defined in a config file.

On **Linux/Mac**, export the following environment variables:

```shell
export SAIL_BASEURL=https://{org}.api.identitynow.com
export SAIL_TOKENURL=https://{org}.api.identitynow.com/oauth/token
export SAIL_CLIENTID={clientID}
export SAIL_CLIENTSECRET={clientSecret}
export SAIL_DEBUG=false
```

If you want your environment variables to persist across terminal sessions, you will need to add these exports to your shell profile, like `~/.bash_profile`.

On **Windows PowerShell** run:

```powershell
$env:SAIL_BASEURL = 'https://{org}.api.identitynow.com'
$env:SAIL_TOKENURL = 'https://{org}.api.identitynow.com/oauth/token'
$env:SAIL_CLIENTID = '{clientID}'
$env:SAIL_CLIENTSECRET = '{clientSecret}'
$env:SAIL_DEBUG = 'false'
```

If you want your environment variables to persist across PowerShell sessions, then use the following command instead:

```powershell
[System.Environment]::SetEnvironmentVariable('SAIL_BASEURL','https://{org}.api.identitynow.com')
```

# Usage

Run the `sail` command for an overview of the available commands and flags.  You can use the `-h` flag with any available command to see additional options available for each command. You can find more information about each command below.

- [connectors](./cmd/connector/README.md)
- [transforms](./cmd/transform/README.md)

# Contributing

Before you contribute you [must sign our CLA](https://cla-assistant.io/sailpoint-oss/sailpoint-cli). Please read our [contribution guidelines](https://github.com/sailpoint-oss/sailpoint-cli/blob/main/CONTRIBUTING.md) to learn how to contribute to this tool.

# Code of Conduct

We pledge to act and interact in ways that contribute to an open, welcoming, diverse, inclusive, and healthy community. Read our [code of conduct](https://github.com/sailpoint-oss/sailpoint-cli/blob/main/CODE_OF_CONDUCT.md) to learn more.
