==Long==
# Update

Update a workflow in IdentityNow.

Arguments can be a list of directories or files. You can update multiple workflows by specifying multiple file paths as arguments.
If a directory is specified, all JSON files in the directory will be parsed and the workflows uploaded.

## API References:
 - https://developer.sailpoint.com/idn/api/beta/update-workflow
====

==Example==
## File:
```bash
sail workflow update -f {file-path} {file-path}
```

## Directory:
```bash
sail workflow update -d {folder-path} {folder-path}
```
====