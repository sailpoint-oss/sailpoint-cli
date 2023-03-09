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

Run the following command to set the current authentication method for the CLI.

> :warning: **Currently only Personal Access Token Authentication is supported**: OAuth possibly coming in the future!

```shell
sail set auth pat
```

This will set the currently active authentication method to PAT.

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
