# SP-Config

The `spconfig` command makes it easy to perform import and export operations in Identity Security Cloud using the SailPoint CLI.

- [SP-Config](#sp-config)
  - [Download](#download)
    - [Command](#command)
      - [Windows](#windows)
      - [MacOS/Linux](#macoslinux)
    - [Flags](#flags)
      - [Import](#import)
      - [Export](#export)
      - [Folder Path](#folder-path)
  - [Export](#export-1)
    - [Command](#command-1)
    - [Flags](#flags-1)
      - [Import](#import-1)
      - [Export](#export-2)
      - [Folder Path](#folder-path-1)

## Download

### Command

Run the following command to download the results of import of export jobs

#### Windows

```shell
sail spconfig download
-export <export job id>
-import <import job id>
```

#### MacOS/Linux

```shell
sail spconfig download \
  -export <export job id> \
  -export <export job id>
```

### Flags

#### Import

Specifies the ids of the import jobs to download

```shell
sail spconfig download \
  -import <import job id> \
  -import <import job id>
```

#### Export

Specifies the ids of the export jobs to download

```shell
sail spconfig download \
  -export <export job id> \
  -export <export job id>
```

#### Folder Path

Specify the folder path to save the search results in

```shell
sail spconfig download \
  -export <export job id> \
  -export <export job id> \
  -folderPath ./local/folder/path
```

## Export

### Command

Run the following command to begin an spconfig export job in Identity Security Cloud

```shell
sail spconfig export \
  --include SOURCE \
  --include WORKFLOW \
  --description "optional description for the export job"
```

### Flags

#### Include

Specifies the object types to include in the export

```shell
sail spconfig export --include SOURCE --include WORKFLOW
```

#### Exclude

Specifies the object types to exclude from the export

```shell
sail spconfig export --exclude SOURCE --exclude WORKFLOW
```

#### Wait

Wait for the export job to complete and automatically download the results

```shell
sail spconfig export --include SOURCE --wait
```

#### Folder Path

Specify the folder path to save the export results in

```shell
sail spconfig export --include SOURCE --wait -f ./local/folder/path
```
