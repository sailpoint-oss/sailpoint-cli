# Set

The `set` command makes it easy to update configuration values for the SailPoint CLI.

- [Set](#set)
  - [Authentication](#authentication)
  - [Debug](#debug)
  - [Export Templates](#export-templates)
    - [Command](#command)
    - [File Format](#file-format)
  - [Search Templates](#search-templates)
    - [Command](#command-1)
    - [File Format](#file-format-1)

## Authentication

The CLI can authenticate to Identity Security Cloud with a personal access token, OAuth, or a bearer token that another system has already minted.

### Personal access token

```shell
sail set auth pat
sail set pat
```

PAT authentication exchanges a client ID and client secret at the OAuth token endpoint. You can also set `SAIL_BASE_URL`, `SAIL_CLIENT_ID`, and `SAIL_CLIENT_SECRET`.

### OAuth

```shell
sail set auth oauth
```

OAuth signs in through the Identity Security Cloud web portal.

### Bearer token

Use bearer authentication when you receive an access token instead of a client ID and client secret. That token is the `access_token` returned after a client ID and client secret are presented to the Identity Security Cloud OAuth token endpoint. The CLI sends it as `Authorization: Bearer <token>` on every request and does not call the OAuth token endpoint again.

For a deployment, or any other process that receives the token from another system:

```shell
export SAIL_BASE_URL="https://<tenant>.api.identitynow.com"
export SAIL_ACCESS_TOKEN="<access_token>"
```

`SAIL_ACCESS_TOKEN` selects bearer authentication for that process, including when another authentication method is configured. You can also pass the token for a single invocation:

```shell
sail --access-token "<access_token>" <command>
```

To store the token for the active environment:

```shell
sail set auth bearer
sail set bearer --token "<access_token>"
```

`sail set bearer` prompts for the token when `--token` is omitted. The value may be the raw access token or the full `Bearer <access_token>` header value. A stored token is kept in the system keyring, not in the config file, and is used for later commands until it expires. The CLI cannot refresh it. Mint a new token at the OAuth token endpoint and set it again.

## Debug

Run the following command to set the debug flag on the CLI.

```shell
sail set debug enable
```

## Export Templates

### Command

Run the following command to populate the path to a custom Export Template JSON file.

```shell
sail set exportTemplates "path/to/export/template/file"

or

sail set export "path/to/export/template/file"
```

### File Format

```json
[
  {
    "name": "all-objects",
    "description": "Export all available objects",
    "variables": [],
    "exportBody": {
      "description": "Export all available objects",
      "excludeTypes": [],
      "includeTypes": [
        "SOURCE",
        "RULE",
        "TRIGGER_SUBSCRIPTION",
        "TRANSFORM",
        "IDENTITY_PROFILE"
      ],
      "objectOptions": {}
    }
  }
]
```

## Search Templates

### Command

Run the following command to populate the path to a custom Search Template JSON file.

```shell
sail set searchTemplates "path/to/search/template/file"

or

sail set search "path/to/search/template/file"
```

### File Format

Below is an example of the search template file format:

- The first template is an example of one using variables in its query
- The second is an example of a fully predefined template with no variables

```json
[
  {
    "name": "all-provisioning-events",
    "description": "All provisioning events in the tenant for a given time range",
    "variables": [{ "name": "days", "prompt": "Days before today" }],
    "searchQuery": {
      "indices": ["events"],
      "queryType": null,
      "queryVersion": null,
      "query": {
        "query": "(type:provisioning AND created:[now-{{days}}d TO now])"
      },
      "sort": [],
      "searchAfter": []
    }
  },
  {
    "name": "all-provisioning-events-90-days",
    "description": "All provisioning events in the tenant for a given time range",
    "variables": [],
    "searchQuery": {
      "indices": ["events"],
      "queryType": null,
      "queryVersion": null,
      "query": {
        "query": "(type:provisioning AND created:[now-90d TO now])"
      },
      "sort": [],
      "searchAfter": []
    }
  }
]
```
