# Search

The `search` command makes it easy to search in identitynow using the SailPoint CLI.

- [Search](#search)
  - [Query](#query)
    - [Command](#command)
    - [Flags](#flags)
      - [Indicies](#indicies)
      - [Sort](#sort)
      - [Output Types](#output-types)
      - [Folder Path](#folder-path)
  - [Template](#template)
    - [Command](#command-1)
    - [Flags](#flags-1)
      - [Output Types](#output-types-1)
      - [Folder Path](#folder-path-1)

## Query

### Command

Run the following command to search with manually provided search attributes.

```shell
sail search query <search query string> -indicies <indicie to search on>
```

### Flags

#### Indicies

Specifies the indicies to run the search operation on

```shell
sail search query "name:a*" -indicies identities
```

You can search multiple indicies by added additional flags

```shell
sail search query "name:a*" -indicies identities -indicies accessprofiles
```

#### Sort

Specified the sort strings used for the search query

```shell
sail search query "name:a*" -indicies identities -sort name -sort "-created"
```

#### Output Types

Specify the output data format currently `json` and `csv` are the only supported types

```shell
sail search query "name:a*" -indicies identities -outputTypes json
```

#### Folder Path

Specify the folder path to save the search results in

```shell
sail search query "name:a*" -indicies identities -folderPath ./local/folder/path
```

## Template

### Command

Run the following command to search with a predefined template.

```shell
sail search template all-provisioning-events-90-days
```

### Flags

#### Output Types

Specify the output data format currently `json` and `csv` are the only supported types

```shell
sail search query "name:a*" -indicies identities -outputTypes json
```

#### Folder Path

Specify the folder path to save the search results in

```shell
sail search query "name:a*" -indicies identities -folderPath ./local/folder/path
```
