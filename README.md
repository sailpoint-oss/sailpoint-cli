[![Discourse Topics][discourse-shield]][discourse-url]
![Times Downloaded][downloads-shield] ![Issues][issues-shield]
![Latest Releases][release-shield] ![Contributor Shield][contributor-shield]
![License Shield][license-shield]

[discourse-shield]:
  https://img.shields.io/discourse/topics?label=Discuss%20This%20Tool&server=https%3A%2F%2Fdeveloper.sailpoint.com%2Fdiscuss
[discourse-url]: https://developer.sailpoint.com/discuss
[downloads-shield]:
  https://img.shields.io/github/downloads/sailpoint-oss/sailpoint-cli/total?label=Downloads
[issues-shield]:
  https://img.shields.io/github/issues/sailpoint-oss/sailpoint-cli?label=Issues
[release-shield]:
  https://img.shields.io/github/v/release/sailpoint-oss/sailpoint-cli?label=Current%20Release
[contributor-shield]:
  https://img.shields.io/github/contributors/sailpoint-oss/sailpoint-cli?label=Contributors
[license-shield]: https://img.shields.io/badge/MIT-License-green

> **CAUTION:** The SailPoint CLI is currently in pre-production and undergoing
> heavy development. Until the CLI reaches version 1.0.0, breaking changes may
> be introduced at any time while we work on refining the CLI.

<!-- PROJECT LOGO -->
<br />
<div align="center">
    <img src="./img/icon.png" alt="Logo">

  <h3 align="center">SailPoint CLI - README</h3>
  <br/>
<div align="center">
<img src="./img/screenshot.png" width="500" height="" style="text-align:center">
</div>
</div>

<!-- ABOUT THE PROJECT -->

## About The Project

The SailPoint Command Line Interface (CLI) makes it easy to interact with
SailPoint's SaaS Platform in a programmatic way. Many functions that use to be
accomplished through tools like Postman or from custom scripts can now be done
directly on the command line with minimal setup.

Please use GitHub
[issues](https://github.com/sailpoint-oss/sailpoint-cli/issues) to
[submit bugs](https://github.com/sailpoint-oss/sailpoint-cli/issues/new?assignees=&labels=&template=bug-report.md&title=%5BBug%5D+Your+Bug+Report+Here)
or make
[feature requests](https://github.com/sailpoint-oss/sailpoint-cli/issues/new?assignees=&labels=&template=feature-request.md&title=%5BFeature%5D+Your+Feature+Request+Here).

If you'd like to contribute directly (which we encourage!) please read the
contribution guidelines below, first!

## Contents

- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [MacOS and Linux](#macos-and-linux)
  - [Windows](#windows)
- [Configuration](#configuration)
  - [Assisted Configuration](#assisted-configuration)
  - [Manual Configuration](#manual-configuration)
  - [Environment Variable Configuration](#environment-variable-configuration)
- [Usage](#usage)
- [Discuss](#discuss)
- [License](#license)
- [Contributing](#contributing)
- [Code of Conduct](#code-of-conduct)

<!-- GETTING STARTED -->

## Getting Started

If you are looking to use the SailPoint CLI, please use the
[Releases](https://github.com/sailpoint-oss/sailpoint-cli/releases) section. If
you want to build this project locally, follow the steps below.

### Prerequisites

- [Golang](https://go.dev/doc/install) version 1.17 or above.

### MacOS and Linux

Open your terminal app, navigate to the project directory, and run the following
command:

```shell
make install
```

After that, make sure you can run the `sail` command.

```shell
sail
```

### Windows

Open PowerShell **as administrator**, navigate to the project directory, and run
the following command.

```bash
go build -o "C:\Program Files\sailpoint\sail.exe"
```

After that, add the following directory to the system PATH parameter. You can
find instructions on how to do this from
[this article](https://medium.com/@kevinmarkvi/how-to-add-executables-to-your-path-in-windows-5ffa4ce61a53).
This will only need to be done the first time you install the CLI.

```
C:\Program Files\sailpoint
```

After setting your environment variable, close all instances of your PowerShell
or Command Prompt, open a new instance, and make sure you can run the `sail`
command.

```shell
sail
```

## Configuration

Before you begin, you will need to gather the following information.

- Create a
  [personal access token](https://developer.sailpoint.com/idn/api/authentication#personal-access-tokens),
  which will be used to authenticate the SailPoint CLI to your IdentityNow
  tenant. Take note of the **Client ID** and the **Client Secret**.
- [Find your org/tenant name](https://developer.sailpoint.com/idn/api/getting-started#find-your-tenant-name).

### Assisted configuration

Run the configure command to configure the CLI for your tenant. This command
will create a configuration file in your home directory to store your tenant's
connection details.

```shell
sail configure
```

### Manual configuration

Alternatively, you can manually create a configuration file in your home
directory.

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
clientSecret: { clientSecret }
clientID: { clientID }
debug: false # Set to true for additional output
```

### Environment variable configuration

You may also specify environment variables for your configuration. This can
useful when using the CLI in an automated environment, like a CI/CD pipeline,
where consuming the configuration from environment variables would be easier
than creating the config file. Environment variables take precedent over values
defined in a config file.

On **Linux/Mac**, export the following environment variables:

```shell
export SAIL_BASEURL=https://{org}.api.identitynow.com
export SAIL_TOKENURL=https://{org}.api.identitynow.com/oauth/token
export SAIL_CLIENTID={clientID}
export SAIL_CLIENTSECRET={clientSecret}
export SAIL_DEBUG=false
```

If you want your environment variables to persist across terminal sessions, you
will need to add these exports to your shell profile, like `~/.bash_profile`.

On **Windows PowerShell** run:

```powershell
$env:SAIL_BASEURL = 'https://{org}.api.identitynow.com'
$env:SAIL_TOKENURL = 'https://{org}.api.identitynow.com/oauth/token'
$env:SAIL_CLIENTID = '{clientID}'
$env:SAIL_CLIENTSECRET = '{clientSecret}'
$env:SAIL_DEBUG = 'false'
```

If you want your environment variables to persist across PowerShell sessions,
then use the following command instead:

```powershell
[System.Environment]::SetEnvironmentVariable('SAIL_BASEURL','https://{org}.api.identitynow.com')
```

## Usage

Run the `sail` command for an overview of the available commands and flags. You
can use the `-h` flag with any available command to see additional options
available for each command. You can find more information about each command
below.

- [connectors](./cmd/connector/README.md)
- [transforms](./cmd/transform/README.md)

## Discuss

[Click Here](https://developer.sailpoint.com/discuss) to discuss this tool with
other users.

<!-- LICENSE -->

## License

Distributed under the MIT License. See [the license](./LICENSE) for more
information.

<!-- CONTRIBUTING -->

## Contributing

Before you contribute you
[must sign our CLA](https://cla-assistant.io/sailpoint-oss/sailpoint-cli).
Please also read our [contribution guidelines](./CONTRIBUTING.md) for all the
details on contributing.

<!-- CODE OF CONDUCT -->

## Code of Conduct

We pledge to act and interact in ways that contribute to an open, welcoming,
diverse, inclusive, and healthy community. Read our
[code of conduct](./CODE_OF_CONDUCT.md) to learn more.

<p align="right">(<a href="#readme-top">back to top</a>)</p>
